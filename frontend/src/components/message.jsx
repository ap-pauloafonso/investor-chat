import React from "react";

const Message = ({ username, message, isSender, isBot }) => {
  return (
    <>
      {isBot ? (
        <div className="flex items-center justify-start">
          <div className="flex gap-1 bg-black text-white rounded-lg p-2">
            <svg
              className={"h-6 w-auto fill-current"}
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
            >
              <title>robot</title>
              <path d="M12,2A2,2 0 0,1 14,4C14,4.74 13.6,5.39 13,5.73V7H14A7,7 0 0,1 21,14H22A1,1 0 0,1 23,15V18A1,1 0 0,1 22,19H21V20A2,2 0 0,1 19,22H5A2,2 0 0,1 3,20V19H2A1,1 0 0,1 1,18V15A1,1 0 0,1 2,14H3A7,7 0 0,1 10,7H11V5.73C10.4,5.39 10,4.74 10,4A2,2 0 0,1 12,2M7.5,13A2.5,2.5 0 0,0 5,15.5A2.5,2.5 0 0,0 7.5,18A2.5,2.5 0 0,0 10,15.5A2.5,2.5 0 0,0 7.5,13M16.5,13A2.5,2.5 0 0,0 14,15.5A2.5,2.5 0 0,0 16.5,18A2.5,2.5 0 0,0 19,15.5A2.5,2.5 0 0,0 16.5,13Z" />
            </svg>
            <strong>{username}:</strong>
            <p>{message}</p>
          </div>
        </div>
      ) : isSender ? (
        <div className="flex items-center justify-end">
          <div className="flex flex-col bg-blue-500 rounded-lg p-2">
            <p className="text-white">{message}</p>
          </div>
        </div>
      ) : (
        <div className="flex items-center justify-start">
          <div className="flex gap-1 bg-gray-200 rounded-lg p-2">
            <strong>{username}:</strong>
            <p className="text-gray-800">{message}</p>
          </div>
        </div>
      )}
    </>
  );
};

export default Message;
