import { atom } from "nanostores"
import { UnaryInterceptor } from "grpc-web";
import { $authToken } from "../auth/store/auth";
import { GetUserByEmailRequest, User, UserServiceClient } from "../proto/userservice"

export const $currentDbUser = atom<User | null>(null)

const unaryInterceptor: UnaryInterceptor<any, any> = {
    intercept: (request, invoker) => {
      const m = request.getMetadata();
      const token = $authToken.get();
      m["authorization"] = token;
      return invoker(request);
    },
};
  
export const userService = new UserServiceClient(
    import.meta.env.VITE_PUBLIC_API_URL.replace(/\/$/, ""),
    {},
    {
        unaryInterceptors: [unaryInterceptor],
    }
);

export const createUserIfNotExists = async (email: string): Promise<User> => {
    try {
        // TODO: Uncomment after running `go generate` to regenerate proto files
        const response = await userService.CreateUserIfNotExists(
            GetUserByEmailRequest.fromObject({ email }),
            {}
        );

        // Temporary fallback - try to get user by email
        // const response = await userService.GetUserByEmail(
        //     GetUserByEmailRequest.fromObject({ email }),
        //     {}
        // );

        $currentDbUser.set(response);
        return response;
    } catch (error) {
        console.error("Error creating/fetching user:", error);
        throw error;
    }
};

export const getUserByEmail = async (email: string): Promise<User> => {
    try {
        const response = await userService.GetUserByEmail(
            GetUserByEmailRequest.fromObject({ email }),
            {}
        );

        return response;
    } catch (error) {
        console.error("Error fetching user by email:", error);
        throw error;
    }
}; 