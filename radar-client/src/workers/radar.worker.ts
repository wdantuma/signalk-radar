/// <reference lib="webworker" />

import { RadarMessage } from './RadarMessage'

const socket = new WebSocket('ws://localhost:3000/radar/v1/stream');
socket.binaryType = "arraybuffer"

socket.onmessage = (event) => {
  let message = RadarMessage.deserialize(event.data)
  //postMessage({angle:message.spoke.angle})    
}

socket.onclose = (event) => {
  console.log("Close")
}

socket.onopen = (event) => {
  console.log("Open")
}

socket.onerror = (event) => {
  console.log("Errot")
}

// addEventListener('message', ({ data }) => {
//   const response = `worker response to ${data}`;
//   postMessage(response);
// });
