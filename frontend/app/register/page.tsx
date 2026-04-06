"use client";
import { useRouter } from "next/navigation";
import { useState } from "react";

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export default function Login() {
	const router = useRouter();
	const [name, setName] = useState("");
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");

	async function handleSubmit(e) {
		e.preventDefault();
		const res = await fetch(`${API_URL}/register`, {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			credentials: "include",
			body: JSON.stringify({ name, email, password }),
		});

		const data = await res.json();
		if (res.ok) {
			console.log(data.message);
			router.push("/login");
		}
	}

	return (
		<>
			<form
				onSubmit={handleSubmit}
				className="max-w-sm mx-auto p-4 border rounded space-y-3"
			>
				<label htmlFor="name" className="block text-sm font-medium">
					Name:
				</label>
				<input
					type="text"
					id="name"
					value={name}
					onChange={(e) => setName(e.target.value)}
					className="w-full border rounded px-2 py-1"
				/>

				<label htmlFor="email" className="block text-sm font-medium">
					Email:
				</label>
				<input
					type="email"
					id="email"
					value={email}
					onChange={(e) => setEmail(e.target.value)}
					className="w-full border rounded px-2 py-1"
				/>

				<label htmlFor="password" className="block text-sm font-medium">
					Password:
				</label>
				<input
					type="password"
					id="password"
					value={password}
					onChange={(e) => setPassword(e.target.value)}
					className="w-full border rounded px-2 py-1"
				/>

				<button
					type="submit"
					className="w-full bg-blue-500 text-white py-2 rounded hover:bg-blue-600"
				>
					Register
				</button>
			</form>
		</>
	);
}
