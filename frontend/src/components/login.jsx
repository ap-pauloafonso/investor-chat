// Login.js

import React, { useState } from "react";
import { toast } from "react-toastify";

function Login() {
  const [loginUser, setLoginUser] = useState({ username: "", password: "" });
  const [signupUser, setSignupUser] = useState({ username: "", password: "" });

  function handleOnChangeLogin(e) {
    setLoginUser({ ...loginUser, [e.target.name]: e.target.value });
  }

  function handleOnChangeSignup(e) {
    setSignupUser({ ...signupUser, [e.target.name]: e.target.value });
  }

  const handleLogin = async () => {
    console.log(JSON.stringify(loginUser));
    const response = await fetch("http://localhost:8080/api/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(loginUser),
    });

    const data = await response.json();

    if (response.status === 200) {
      toast.success("Login successful", {
        position: "top-right",
        autoClose: 5000, // Close after 5 seconds
      });
    } else {
      toast.error(data.errorMessage, {
        position: "top-right",
        autoClose: 5000, // Close after 5 seconds
      });
    }
  };

  const handleSignup = async () => {
    const response = await fetch("http://localhost:8080/api/register", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(signupUser),
    });

    if (response.status === 200) {
      toast.success("Register successful", {
        position: "top-right",
        autoClose: 5000, // Close after 5 seconds
      });
    } else {
      const data = await response.json();

      console.log();
      toast.error(data.errorMessage, {
        position: "top-right",
        autoClose: 5000, // Close after 5 seconds
      });
    }
  };

  return (
    <div className={"min-h-screen bg-gray-100 flex"}>
      <div className="container flex  justify-center items-center self-center">
        <div className="max-w-[400px] w-full m-auto p-6 bg-white dark:bg-slate-900 shadow-md dark:shadow-gray-800 rounded-md">
          <h5 className="my-6 text-xl font-semibold">Login</h5>
          <form className="text-start" data-bitwarden-watching="1">
            <div className="grid grid-cols-1">
              <div className="mb-4">
                <label className="font-semibold" for="LoginEmail">
                  Login:
                </label>
                <input
                  onChange={(e) => handleOnChangeLogin(e)}
                  name={"username"}
                  class="form-input mt-3 w-full py-2 px-3 h-10 bg-transparent dark:bg-slate-900 dark:text-slate-200 rounded outline-none border border-gray-200 focus:border-indigo-600 dark:border-gray-800 dark:focus:border-indigo-600 focus:ring-0"
                  placeholder="name@example.com"
                />
              </div>

              <div className="mb-4">
                <label className="font-semibold" for="LoginPassword">
                  Password:
                </label>
                <input
                  onChange={(e) => handleOnChangeLogin(e)}
                  name={"password"}
                  type="password"
                  class="form-input mt-3 w-full py-2 px-3 h-10 bg-transparent dark:bg-slate-900 dark:text-slate-200 rounded outline-none border border-gray-200 focus:border-indigo-600 dark:border-gray-800 dark:focus:border-indigo-600 focus:ring-0"
                  placeholder="Password:"
                />
              </div>

              <div className="mb-4">
                <input
                  onClick={handleLogin}
                  type="button"
                  class="py-2 px-5 inline-block tracking-wide border align-middle duration-500 text-base text-center bg-green-600 hover:bg-green-700 border-green-600 hover:border-green-700 text-white rounded-md w-full"
                  value="Login"
                />
              </div>
            </div>
          </form>
        </div>

        <div className="max-w-[400px] w-full m-auto p-6 bg-white dark:bg-slate-900 shadow-md dark:shadow-gray-800 rounded-md">
          <h5 className="my-6 text-xl font-semibold">Sign up</h5>
          <form className="text-start" data-bitwarden-watching="1">
            <div className="grid grid-cols-1">
              <div className="mb-4">
                <label className="font-semibold">Login:</label>
                <input
                  onChange={(e) => handleOnChangeSignup(e)}
                  name={"username"}
                  className="form-input mt-3 w-full py-2 px-3 h-10 bg-transparent dark:bg-slate-900 dark:text-slate-200 rounded outline-none border border-gray-200 focus:border-indigo-600 dark:border-gray-800 dark:focus:border-indigo-600 focus:ring-0"
                  placeholder="name@example.com"
                />
              </div>

              <div className="mb-4">
                <label className="font-semibold">Password:</label>
                <input
                  onChange={(e) => handleOnChangeSignup(e)}
                  name={"password"}
                  type="password"
                  className="form-input mt-3 w-full py-2 px-3 h-10 bg-transparent dark:bg-slate-900 dark:text-slate-200 rounded outline-none border border-gray-200 focus:border-indigo-600 dark:border-gray-800 dark:focus:border-indigo-600 focus:ring-0"
                  placeholder="Password:"
                />
              </div>

              <div className="mb-4">
                <input
                  onClick={handleSignup}
                  type={"button"}
                  className="py-2 px-5 inline-block tracking-wide border align-middle duration-500 text-base text-center bg-indigo-600 hover:bg-indigo-700 border-indigo-600 hover:border-indigo-700 text-white rounded-md w-full"
                  value="Sign up"
                />
              </div>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

export default Login;
