// src/components/Chat.js
import React, { useEffect, useRef, useState } from "react";
import { toast } from "react-toastify";
import Message from "./message";
import clsx from "clsx";

const Chat = ({ userName, logoutFn }) => {
  const [channels, setChannels] = useState([]);
  const [newChannel, setNewChannel] = useState("");

  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState("");
  const [socket, setSocket] = useState(null);

  const [selectedChannel, setSelectedChannel] = useState("default");

  const el = useRef(null);

  function scrollToBottom() {
    el.current.scrollIntoView({ block: "end", behavior: "smooth" });
  }

  const connectWebSocket = (channel) => {
    if (socket) {
      socket.close();
    }

    var wsUrl = `ws://localhost:8080/ws/${selectedChannel}`;
    const ws = new WebSocket(wsUrl);
    setSocket(ws);

    ws.onopen = () => {
      console.log("Connected to the WebSocket server");
    };

    ws.onmessage = (event) => {
      const message = event.data;

      if (message === "[channel_list_update]") {
        fetchChannels();
        return;
      }
      const obj = JSON.parse(message);
      if (Array.isArray(obj)) {
        const mappedArray = obj.map((x) => ({
          msg: x.Msg,
          user: x.Username,
          isBot: x.IsBot,
        }));

        setMessages(mappedArray);
      } else {
        setMessages((prevMessages) => [
          ...prevMessages,
          { msg: obj.Msg, user: obj.Username, isBot: obj.IsBot },
        ]);
      }

      scrollToBottom();
    };

    ws.onclose = (event) => {
      if (event.wasClean) {
        return; // no need for reconnection
      }
      console.log("attempting to reconnect in 2s...");
      setTimeout(connectWebSocket, 2000);
    };
  };

  useEffect(() => {
    fetchChannels();

    return () => {};
  }, []);

  useEffect(() => {
    setNewMessage("");
    // Initial WebSocket connection
    connectWebSocket(selectedChannel);

    return () => {
      console.log("cleanup");
      // Clean up when the component is unmounted
      if (socket) {
        socket.close();
      }
    };
  }, [selectedChannel]);

  const sendMessage = () => {
    if (newMessage.trim() === "") {
      return;
    }

    // Send a message to the WebSocket server
    socket.send(newMessage);
    setNewMessage("");
  };

  function fetchChannels() {
    fetch("http://localhost:8080/api/channels")
      .then((x) => x.json())
      .then((data) => setChannels(data.channels));
  }

  // Function to create a new channel
  const createChannel = async () => {
    if (newChannel.trim() !== "") {
      const response = await fetch("http://localhost:8080/api/channels", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ name: newChannel }),
      });
      if (response.ok) {
        toast.success("channel created", {
          position: "top-right",
          autoClose: 5000, // Close after 5 seconds
        });
        fetchChannels();
      } else {
        var err = await response.json();
        toast.error(err.errorMessage, {
          position: "top-right",
          autoClose: 5000, // Close after 5 seconds
        });
      }

      setNewChannel("");
    }
  };
  function changeChannel(channel) {
    if (channel === selectedChannel) {
      return;
    }

    setSelectedChannel(channel);
  }

  return (
    <div className="flex flex-col h-screen bg-gray-100 gap-3">
      <nav className="bg-blue-500 py-3">
        <div className="container  mx-auto flex justify-between items-center">
          <div className="navbar-left">Chat</div>
          <div className="navbar-right flex items-center space-x-4">
            <div className="text-white">
              Welcome, <strong>{userName}</strong>
            </div>
            <button
              onClick={logoutFn}
              className="bg-blue-700 text-white px-4 py-2 rounded"
            >
              Logout
            </button>
          </div>
        </div>
      </nav>

      <div className="flex container">
        {/* First Column: List of Channels with Add Channel Option (30%) */}
        <div className="flex flex-col min-w-[20%] p-4">
          <h2 className="text-xl font-bold mb-4">Channels</h2>
          <ul className={"overflow-y-scroll max-h-[350px]"}>
            {channels.map((channel, index) => (
              <li
                key={index}
                className={clsx(
                  "mb-2 flex gap-2",
                  channel !== selectedChannel && "cursor-pointer",
                )}
                onClick={() => changeChannel(channel)}
              >
                <span> {channel}</span>

                {channel === selectedChannel && (
                  <div className={"text-green-400"}>
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className={"fill-current h-6"}
                      viewBox="0 0 24 24"
                    >
                      <title>connection</title>
                      <path d="M21.4 7.5C22.2 8.3 22.2 9.6 21.4 10.3L18.6 13.1L10.8 5.3L13.6 2.5C14.4 1.7 15.7 1.7 16.4 2.5L18.2 4.3L21.2 1.3L22.6 2.7L19.6 5.7L21.4 7.5M15.6 13.3L14.2 11.9L11.4 14.7L9.3 12.6L12.1 9.8L10.7 8.4L7.9 11.2L6.4 9.8L3.6 12.6C2.8 13.4 2.8 14.7 3.6 15.4L5.4 17.2L1.4 21.2L2.8 22.6L6.8 18.6L8.6 20.4C9.4 21.2 10.7 21.2 11.4 20.4L14.2 17.6L12.8 16.2L15.6 13.3Z" />
                    </svg>
                  </div>
                )}
              </li>
            ))}
          </ul>
          <div className="mt-2 mb-4 relative">
            <input
              type="text"
              value={newChannel}
              onChange={(e) => setNewChannel(e.target.value)}
              placeholder="New Channel Name"
              className="w-full p-2 rounded-full border border-gray-300 focus:outline-none"
            />
            <button
              onClick={createChannel}
              className="absolute px-4 right-0 top-0 h-10 bg-blue-500 text-white p-2 rounded-full hover:bg-blue-600 transition duration-200"
            >
              Add
            </button>
          </div>
        </div>

        <div className="flex-grow flex w-8/10 p-4  rounded-lg flex-col gap-2 py-2">
          <div
            className={"bg-white h-[400px] rounded-lg overflow-y-scroll py-1"}
          >
            <div
              className={
                "h-full overflow-y-scroll px-[5%] flex flex-col gap-1 pt-2"
              }
            >
              {messages.map((message, index) => (
                <Message
                  key={index}
                  username={message.user}
                  isSender={message.user === userName}
                  message={message.msg}
                  isBot={message.isBot}
                />
              ))}
              <div id={"el"} ref={el}></div>
            </div>
          </div>
          <div className="mb-4 relative">
            <input
              type="text"
              value={newMessage}
              onChange={(e) => setNewMessage(e.target.value)}
              placeholder="Type a message"
              className="w-full p-2 rounded-full border border-gray-300 focus:outline-none"
            />
            <button
              onClick={sendMessage}
              className="absolute px-4 right-0 top-0 h-10 bg-blue-500 text-white p-2 rounded-full hover:bg-blue-600 transition duration-200"
            >
              Send
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Chat;
