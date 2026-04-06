"use client";
import { useEffect, useRef, useState, useCallback } from "react";
import { useRouter } from "next/navigation";

const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";
const WS_URL = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8000/ws";

type Message = {
    chat_id: number;
    sender_name: string;
    content: string;
    sender_id?: number;
    id?: number; // Added ID for React keys
};

export default function Home() {
    const [input, setInput] = useState("");
    const [messages, setMessages] = useState<Message[]>([]);
    const [username, setUsername] = useState("");
    const [chats, setChats] = useState<number[]>([]);
    const [chatsNames, setChatsNames] = useState<Record<string, string[]>>({});
    const [currentChatId, setCurrentChatId] = useState<number | null>(null);
    const [isConnected, setIsConnected] = useState(false);
    const [isLoadingHistory, setIsLoadingHistory] = useState(false); // New loading state

    const ws = useRef<WebSocket | null>(null);
    const router = useRouter();

    // 1. Fetch User Data & Chat List
    useEffect(() => {
        async function getUserData() {
            try {
                const res = await fetch(`${API_URL}`, {
                    method: "GET",
                    credentials: "include",
                });
                if (!res.ok) {
                    router.push("/login");
                    return;
                }
                const data = await res.json();
                setUsername(data.username);
                setChats(data.chats);
                setChatsNames(data.chatsNames);
            } catch (err) {
                console.error("Error fetching user data:", err);
                router.push("/login");
            }
        }
        getUserData();
    }, [router]);

    // 2. WebSocket Connection
    useEffect(() => {
        console.log("Connecting to Global WebSocket...");
        const socket = new WebSocket(WS_URL);

        socket.onopen = () => {
            console.log("WebSocket Connected");
            setIsConnected(true);
        };

        socket.onclose = () => {
            console.log("WebSocket Disconnected");
            setIsConnected(false);
        };

        socket.onerror = (error) => {
            console.error("WebSocket Error:", error);
        };

        ws.current = socket;

        return () => {
            socket.close();
        };
    }, []);

    // 3. NEW: Fetch Message History when Chat Changes
    useEffect(() => {
        if (!currentChatId) return;

        let isMounted = true; // Prevents race conditions if user clicks fast

        async function fetchHistory() {
            setIsLoadingHistory(true);
            try {
                // Ensure your Go backend has this route: GET /chats/:id/messages
                const res = await fetch(
                    `${API_URL}/chats/${currentChatId}/messages`,
                    {
                        method: "GET",
                        credentials: "include",
                    },
                );

                if (res.ok && isMounted) {
                    const history = await res.json();
                    // Ensure history is an array before setting
                    setMessages(Array.isArray(history) ? history : []);
                }
            } catch (err) {
                console.error("Failed to load chat history", err);
            } finally {
                if (isMounted) setIsLoadingHistory(false);
            }
        }

        fetchHistory();

        return () => {
            isMounted = false;
        };
    }, [currentChatId]);

    // 4. WebSocket Message Listener
    useEffect(() => {
        if (!ws.current) return;

        ws.current.onmessage = (event) => {
            try {
                const msg: Message = JSON.parse(event.data);

                // Only append if the message belongs to the active chat
                if (
                    currentChatId &&
                    Number(msg.chat_id) === Number(currentChatId)
                ) {
                    setMessages((prev) => [...prev, msg]);
                }
            } catch (err) {
                console.error("Error parsing message:", err);
            }
        };
    }, [currentChatId]);

    // 5. Chat Selection
    const selectChat = (chatId: number) => {
        // Clear messages immediately to give visual feedback
        setMessages([]);
        setCurrentChatId(chatId);
    };

    // 6. Send Message
    const sendMsg = useCallback(() => {
        if (!ws.current || !isConnected || !currentChatId) return;
        if (!input.trim()) return;

        const payload = {
            chat_id: currentChatId,
            content: input,
        };

        ws.current.send(JSON.stringify(payload));
        setInput("");

        // Optional: Optimistic UI (add message immediately)
        // const optimisticMsg = { ...payload, sender_name: username, chat_id: currentChatId };
        // setMessages(prev => [...prev, optimisticMsg]);
    }, [input, currentChatId, isConnected, username]);

    const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
        if (e.key === "Enter") sendMsg();
    };

    async function logout() {
        if (ws.current) ws.current.close();
        await fetch(`${API_URL}/logout`, {
            method: "GET",
            credentials: "include",
        });
        router.push("/login");
    }

    return (
        <div className="min-h-screen bg-gray-100 p-8">
            <div className="max-w-6xl mx-auto bg-white rounded-lg shadow-lg overflow-hidden h-[80vh] flex flex-col md:flex-row">
                {/* SIDEBAR */}
                <div className="w-full md:w-1/4 bg-gray-50 border-r flex flex-col">
                    <div className="p-4 border-b bg-white">
                        <h2 className="font-bold text-lg text-gray-800 truncate">
                            {username}
                        </h2>
                        <div className="flex items-center gap-2 mt-1">
                            <span
                                className={`h-2.5 w-2.5 rounded-full ${isConnected ? "bg-green-500" : "bg-red-500"}`}
                            ></span>
                            <span className="text-xs text-gray-500">
                                {isConnected ? "Online" : "Offline"}
                            </span>
                        </div>
                    </div>

                    <div className="flex-1 overflow-y-auto p-2 space-y-2">
                        {chats.length > 0 ? (
                            chats.map((chatId) => (
                                <button
                                    key={chatId}
                                    onClick={() => selectChat(chatId)}
                                    className={`w-full text-left px-4 py-3 rounded-lg transition-all duration-200 ${
                                        currentChatId === chatId
                                            ? "bg-blue-600 text-white shadow-md"
                                            : "hover:bg-gray-200 text-gray-700"
                                    }`}
                                >
                                    <div className="font-medium">
                                        {chatsNames[chatId]?.join(", ") ||
                                            `Chat ${chatId}`}
                                    </div>
                                    <div className="text-xs opacity-75 truncate">
                                        Click to open
                                    </div>
                                </button>
                            ))
                        ) : (
                            <div className="text-center text-gray-400 mt-10 text-sm">
                                No chats found
                            </div>
                        )}
                    </div>

                    <div className="p-4 border-t bg-gray-50">
                        <button
                            onClick={logout}
                            className="w-full text-center text-red-500 hover:text-red-700 text-sm font-semibold"
                        >
                            Log out
                        </button>
                    </div>
                </div>

                {/* CHAT AREA */}
                <div className="flex-1 flex flex-col bg-white">
                    {currentChatId ? (
                        <>
                            <div className="p-4 border-b flex justify-between items-center bg-gray-50">
                                <h3 className="font-semibold text-gray-700">
                                    {chatsNames[currentChatId]?.join(", ") ||
                                        `Chat ${currentChatId}`}
                                </h3>
                            </div>

                            <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-white">
                                {isLoadingHistory ? (
                                    <div className="h-full flex items-center justify-center text-gray-400">
                                        Loading history...
                                    </div>
                                ) : messages.length > 0 ? (
                                    messages.map((msg, idx) => {
                                        const isMe =
                                            msg.sender_name === username;
                                        return (
                                            <div
                                                key={idx}
                                                className={`flex ${isMe ? "justify-end" : "justify-start"}`}
                                            >
                                                <div
                                                    className={`max-w-[70%] px-4 py-2 rounded-xl text-sm ${
                                                        isMe
                                                            ? "bg-blue-600 text-white rounded-br-none"
                                                            : "bg-gray-100 text-gray-800 rounded-bl-none"
                                                    }`}
                                                >
                                                    <div className="break-words">
                                                        {msg.content}
                                                    </div>
                                                    {!isMe && (
                                                        <div className="text-xs text-gray-400 mt-1 font-medium">
                                                            {msg.sender_name}
                                                        </div>
                                                    )}
                                                </div>
                                            </div>
                                        );
                                    })
                                ) : (
                                    <div className="h-full flex flex-col items-center justify-center text-gray-300">
                                        <p>No messages yet</p>
                                        <p className="text-xs">
                                            Say hello to start the conversation!
                                        </p>
                                    </div>
                                )}
                            </div>

                            <div className="p-4 border-t bg-gray-50">
                                <div className="flex gap-2">
                                    <input
                                        type="text"
                                        value={input}
                                        onChange={(e) =>
                                            setInput(e.target.value)
                                        }
                                        onKeyPress={handleKeyPress}
                                        className="flex-1 border border-gray-300 rounded-full px-4 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                                        placeholder="Type a message..."
                                        disabled={!isConnected}
                                    />
                                    <button
                                        onClick={sendMsg}
                                        disabled={!isConnected || !input.trim()}
                                        className="bg-blue-600 hover:bg-blue-700 text-white rounded-full px-6 py-2 font-medium disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                                    >
                                        Send
                                    </button>
                                </div>
                            </div>
                        </>
                    ) : (
                        <div className="flex-1 flex items-center justify-center bg-gray-50 text-gray-400 flex-col">
                            <div className="text-xl font-semibold mb-2">
                                Select a chat
                            </div>
                            <p className="text-sm">
                                Choose a conversation from the sidebar to start
                                messaging.
                            </p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
