import { $isLoggedIn, $user, LoginUser, LogoutUser } from "../store/user"
import { useStore } from '@nanostores/react'
import 'firebaseui/dist/firebaseui.css';
import authPlugin from "../plugin/AuthPluginWeb"
import { Redirect, useHistory, useLocation } from 'react-router';
import { useEffect } from 'react';
import { Capacitor } from '@capacitor/core';
import { LocationState } from '../components/ProtectedRoute';

const LoginPage = () => {

  const user = useStore($user);
  const isLoggedIn = useStore($isLoggedIn);
  const location = useLocation()
  const history = useHistory()

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
    <IonPage>

      <IonHeader>
        <IonToolbar>
          <IonButtons slot="start">
            <IonMenuButton />
          </IonButtons>
          <IonTitle>Login</IonTitle>
        </IonToolbar>
      </IonHeader>

        <div style={{ textAlign: 'center' }}>
          <h2>Proof Track</h2>
        </div>   

          {/* Redirect to next page */}
          {isLoggedIn?<Redirect to="/dashboard" />:<></>}
          

          {
            isLoggedIn ?
              <>
                <IonButton onClick={() => { LogoutUser() }}>Logout</IonButton>
              </>
              :
              <IonButton onClick={() => {

                const _user = authPlugin.login({ value: '' })
                // todo: handle better
                _user.then((user) => {
                  console.debug("LoginPage: user recieved")
                  authPlugin.getToken().then((token) => {
                    LoginUser(user.user, token.token)
                  })

                  if(from)
                    history.replace(from);

                })
              }}>Show Login</IonButton>
          }

      <div id="firebaseui-auth-container"></div>

      </IonPage>
  )
}

export default LoginPage

