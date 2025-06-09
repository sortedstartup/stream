import { atom, onMount } from "nanostores"
import { UnaryInterceptor } from "grpc-web";
import { $authToken } from "../auth/store/auth";
import { 
    GetVideoRequest, 
    ListVideosRequest, 
    Video, 
    VideoServiceClient,
    AddUserToSpaceRequest,
    RemoveUserFromSpaceRequest,
    ListSpaceMembersRequest,
    ListUsersRequest,
    UpdateUserSpaceAccessRequest,
    AccessLevel,
    SpaceMember,
    User
} from "../proto/videoservice"

export const $videos = atom<Video[]>([])

onMount($videos,() => {
    console.log("videos.ts -> onMount()")
    fetchVideos()
})

const unaryInterceptor: UnaryInterceptor<any, any> = {
    intercept: (request, invoker) => {
      const m = request.getMetadata();
      const token = $authToken.get();
      m["authorization"] = token; //`${$authContext.get().user.token}`;
      return invoker(request);
    },
  };
  
export const videoService = new VideoServiceClient(
    import.meta.env.VITE_PUBLIC_API_URL.replace(/\/$/, ""),
    {},
    {
        unaryInterceptors: [unaryInterceptor],
    }
);

export const fetchVideos = async () => {
    const response = await videoService.ListVideos(ListVideosRequest.fromObject({
        pageNumber: 0,
        pageSize: 10,
    }),{})

    $videos.set(response.videos)
}

export const fetchVideo = async (id: string) => {
    const response = await videoService.GetVideo(GetVideoRequest.fromObject({
         video_id: id
    }),{})

    return response
}

// Space sharing methods
export const addUserToSpace = async (spaceId: string, userId: string, accessLevel: AccessLevel) => {
    const response = await videoService.AddUserToSpace(AddUserToSpaceRequest.fromObject({
        space_id: spaceId,
        user_id: userId,
        access_level: accessLevel
    }), {})
    return response
}

export const removeUserFromSpace = async (spaceId: string, userId: string) => {
    const response = await videoService.RemoveUserFromSpace(RemoveUserFromSpaceRequest.fromObject({
        space_id: spaceId,
        user_id: userId
    }), {})
    return response
}

export const listSpaceMembers = async (spaceId: string): Promise<SpaceMember[]> => {
    const response = await videoService.ListSpaceMembers(ListSpaceMembersRequest.fromObject({
        space_id: spaceId
    }), {})
    return response.members
}

export const listUsers = async (): Promise<User[]> => {
    const response = await videoService.ListUsers(ListUsersRequest.fromObject({}), {})
    return response.users
}

export const updateUserSpaceAccess = async (spaceId: string, userId: string, accessLevel: AccessLevel) => {
    const response = await videoService.UpdateUserSpaceAccess(UpdateUserSpaceAccessRequest.fromObject({
        space_id: spaceId,
        user_id: userId,
        access_level: accessLevel
    }), {})
    return response
}