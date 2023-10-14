import React, { useEffect, useState } from "react";
import ReactDOM from "react-dom";
import Login from "./components/login";
import Chat from "./components/chat";

import "./styles.css";
import { ToastContainer } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";

import { CookiesProvider, useCookies } from "react-cookie";
import jwtDecode from "jwt-decode";

function App() {
  const [cookies, setCookie, removeCookie] = useCookies(["token"]);

  const [currentUser, setCurrentUser] = useState("");

  useEffect(() => {
    if (cookies.token) {
      setCurrentUser(jwtDecode(cookies["token"]).username);
    }
  }, [cookies]);

  function handleLogout() {
    setCookie("token", "", { path: "/", expires: new Date(0) });
  }

  return (
    <>
      {cookies.token ? (
        <Chat logoutFn={handleLogout} userName={currentUser} />
      ) : (
        <Login />
      )}

      <ToastContainer
        pauseOnFocusLoss={false}
        pauseOnHover={false}
        theme="colored"
      />
    </>
  );
}

ReactDOM.render(<App />, document.getElementById("root"));
