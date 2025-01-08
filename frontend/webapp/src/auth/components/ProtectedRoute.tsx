// src/components/ProtectedRoute.tsx
import React from 'react';
import { Redirect, Route, RouteProps } from 'react-router-dom';
import { $authToken, $isLoggedIn } from '../store/user';
import { useStore } from '@nanostores/react';

interface ProtectedRouteProps extends RouteProps {
  component: React.ComponentType<any>;
}

export interface LocationState {
  from: {
    pathname: string;
  };
}


const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ component: Component, ...rest }) => {

  const isLoggedIn = useStore($isLoggedIn);
  const authToken = useStore($authToken);

  return (
    <Route {...rest} render={(props) => {
      if (authToken === "INIT") {
        console.debug("ProtectedRoute.tsx: authToken INIT detected waiting...")
        return <div>Verifying Authentication...</div>
      } else {
        console.debug("ProtectedRoute.tsx: valid auth token found, hence render or redirect")
        if (isLoggedIn) {
          return <Component {...props} />
        } else {
          // This state: {from: props } causes problems with the redirect results in blank page
          var redirectPath = null;  
          if(props.location.state!=null) {
              redirectPath = (props.location.state as LocationState).from
          } else {
              redirectPath = props.location;
          }
          return <Redirect to={{ pathname: '/public/login', state: { from: redirectPath } }} />
        }
      }
    }
    } />
  );

}

export default ProtectedRoute;