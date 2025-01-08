import { atom } from "nanostores";
import { User, ANONYMOUS} from "../model/user"
import { persistentAtom } from "@nanostores/persistent";

const jsonSerDe ={
    encode: JSON.stringify,
    decode: JSON.parse,
}

const $user = persistentAtom<User>('user',ANONYMOUS, jsonSerDe)
// const $authToken = persistentAtom<string>('token','INIT', jsonSerDe)
const $authToken = atom<string>('INIT')
const $isLoggedIn = persistentAtom<boolean>('isLoggedIn', false, jsonSerDe)

const LoginUser = (user:User, token:string) => {
    $user.set(user)
    $authToken.set(token)
    $isLoggedIn.set(true)
}

const LogoutUser = () => {
   $user.set(ANONYMOUS)
   $authToken.set("")
    $isLoggedIn.set(false)
}

export {$user, $authToken, $isLoggedIn, LoginUser, LogoutUser}
