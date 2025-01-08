/* eslint-disable object-shorthand */
import { WebPlugin } from '@capacitor/core';
import * as firebaseui from 'firebaseui';
import firebaseApp from 'firebase/compat/app';

import type { AuthPlugin } from '../../plugin/AuthPlugin';
// import { environment } from '../../../environments/environment';
import { User } from '../model/user';

// import { promises } from 'dns';
import { fbapp, fbui } from '../firebase'
// /import { $authContext } from '../stores/user';
const app = fbapp
const ui = fbui//new firebaseui.auth.AuthUI(getAuth());
const auth = fbapp.auth()

export class AuthPluginWeb extends WebPlugin implements AuthPlugin {

    waitForAuth(): Promise<void> {
        // TODO: i dont fully understand how it works
        return new Promise((resolve, reject) => {
            const unsubscribe = auth
                .onAuthStateChanged(
                    (user) => {
                        console.debug("AuthPluginWeb.waitForAuth: onAuthStateChanged")
                        unsubscribe();
                        resolve();
                    },
                    reject // pass up any errors attaching the listener
                );
        });
    }

    getToken(): Promise<{ token: string }> {

        const promise = new Promise<{ token: string }>((resolve, reject) => {

            // check if firebase has fully loaded
            if (auth.currentUser == null) {
                reject("No user logged in")
            }
            auth.currentUser?.getIdToken().then((token) => {
                resolve({ token: token })
            }).catch((err) => {
                console.error(err)
                reject(err)
            })

        });

        return promise;
    }

    logout(): Promise<void> {
        return auth.signOut();
    }

    getUser(): Promise<{ user: User }> {
        return new Promise<{ user: User }>(r => {

            r({ user: new User(auth.currentUser?.displayName as string, auth.currentUser?.email as string, "") })
        });
    }

    getUserRoles(): Promise<Array<string>> {
        return new Promise<Array<string>>(r => {
            r([])
        });
    }

    isLoggedIn(): Promise<{ loggedIn: boolean }> {
        const p = new Promise<{ loggedIn: boolean }>(r => {
            r({ loggedIn: auth.currentUser != null })
        });
        return p;
    }

    getRolesFromToken(): Array<string> {
        return [];
    }

    login(options: { value: string }): Promise<{ user: User }> {

        const promise = new Promise<{ user: User }>((resolve) => {

            // if($authContext.get().isLoggedIn) resolve({user:$authContext.get().user})

            const uiConfig: firebaseui.auth.Config = {

                signInFlow: 'popup',
                signInOptions: [
                    firebaseApp.auth.GoogleAuthProvider.PROVIDER_ID
                ],
                callbacks: {
                    signInSuccessWithAuthResult: function (authResult: any, redirectURL: string) {
                        console.debug(authResult)
                        const fbUser = firebaseApp.auth().currentUser

                        fbUser?.getIdToken().then((token) => {
                            const u = new User(fbUser?.displayName as string, fbUser?.email as string, token)
                            resolve({ user: u })
                        })

                        return false
                    }
                }
            }

            fbui.start('#firebaseui-auth-container', uiConfig)
            // ui.start('#firebaseui-auth-container', {
            //   callbacks: {
            //     // eslint-disable-next-line prefer-arrow/prefer-arrow-functions,
            //     signInSuccessWithAuthResult: function(authResult: any, redirectUrl: any) {
            //         const signedInUser = getAuth().currentUser;
            //         if(currentUser!=null) {
            //             currentUser.getIdTokenResult().then((token)=>{
            //                  // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
            //                 resolve ({user:new User(signedInUser.email!, token.claims.roles as [])});
            //             }).catch((err)=>{
            //                  // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
            //                 resolve({user:new User(signedInUser.email!, [])});
            //             });
            //         }
            //       return false;
            //     },
            //     // eslint-disable-next-line prefer-arrow/prefer-arrow-functions
            //     uiShown: function() {
            //     }
            //   },
            //   signInFlow: 'popup',
            //   signInOptions: [
            //     GoogleAuthProvider.PROVIDER_ID,
            //     GithubAuthProvider.PROVIDER_ID,
            //   ],
            //   tosUrl: '<your-tos-url>',
            //   privacyPolicyUrl: '<your-privacy-policy-url>'
            // });

        });
        return promise;
    }
}
