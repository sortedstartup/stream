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

export const createUserIfNotExists = async (): Promise<User> => {
    try {
        const response = await userService.CreateUserIfNotExists(
            GetUserByEmailRequest.fromObject({}),
            {}
        );

        $currentDbUser.set(response);
        return response;
    } catch (error) {
        console.error("Error creating/fetching user:", error);
        throw error;
    }
}; 