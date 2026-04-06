let ws = new WebSocket("ws://" + location.host + "/ws");

ws.onmessage = (e) => {
    let chat = document.getElementById("chat");
    const data = JSON.parse(e.data);
    chat.innerHTML += "<div>" + `${data.sender}: ${data.content}` + "</div>";
};

function sendMsg() {
    let input = document.getElementById("msg");
    ws.send(input.value);
    input.value = "";
}
