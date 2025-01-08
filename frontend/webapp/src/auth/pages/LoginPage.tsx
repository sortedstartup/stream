import { $isLoggedIn, $user, LoginUser, LogoutUser } from "../store/user"
import { useStore } from '@nanostores/react'
import 'firebaseui/dist/firebaseui.css';
import { AuthPluginWeb } from "../plugin/AuthPluginWeb"
import { useEffect } from 'react';
import { LocationState } from '../components/ProtectedRoute';
import React from 'react';
import { useNavigate, useLocation } from 'react-router';

const LoginPage = () => {
  const user = useStore($user);
  const isLoggedIn = useStore($isLoggedIn);
  const location = useLocation();
  const navigate = useNavigate();
  const authPlugin = new AuthPluginWeb();

  let from = null;
  if(location.state) {
    if((location.state as LocationState).from) {
      from = (location.state as LocationState).from
      console.log("-------------------- from: ", from)
    }
  }

  const RedirectToNext = () => {


    console.log("------------- from: ", location)
    /*
     after login :
    */
    if (isLoggedIn ) {

      // no wait needed
      if (1==1 /* came from referrer */) {
      }
    }
    return ""
   
  }

  return (
    <div className="page">
      <header>
        <nav>
          <button className="menu-button">☰</button>
          <h1>Login</h1>
        </nav>
      </header>

      <div style={{ textAlign: 'center' }}>
        <h2>Proof Track</h2>
      </div>   

      {/* Redirect to next page */}
      {isLoggedIn && navigate("/dashboard")}

      {
        isLoggedIn ? (
          <button onClick={() => { LogoutUser() }}>Logout</button>
        ) : (
          <button onClick={() => {
            const _user = authPlugin.login({ value: '' })
            // todo: handle better
            _user.then((user) => {
              console.debug("LoginPage: user recieved")
              authPlugin.getToken().then((token) => {
                LoginUser(user.user, token.token)
              })

              if(from) {
                navigate(from);
              }
            })
          }}>Show Login</button>
        )
      }

      <div id="firebaseui-auth-container"></div>
    </div>
  )
}

export default LoginPage

