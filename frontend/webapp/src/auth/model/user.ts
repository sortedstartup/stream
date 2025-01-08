class User {
    name: string = ''
    email: string = ''
    token: string = ''

    constructor(name: string, email: string, token: string) {
        this.name = name
        this.email = email
        this.token = token
    }
}

const ANONYMOUS: User = new User("Anonymous", "anonymous@example.com", "")

class AuthContext {
    isLoggedIn: boolean
    user: User
    token: string

    constructor(user: User, isLoggedIn: boolean, token: string) {
        this.user = user
        this.isLoggedIn = isLoggedIn
        this.token = token
    }
}

export { User, ANONYMOUS, AuthContext }

