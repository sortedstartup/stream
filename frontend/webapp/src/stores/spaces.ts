import { atom, onMount } from "nanostores"
import { UnaryInterceptor } from "grpc-web";
import { $authToken } from "../auth/store/auth";
import { 
    CreateSpaceRequest, 
    GetSpaceRequest, 
    ListSpacesRequest, 
    ListVideosInSpaceRequest,
    AddVideoToSpaceRequest,
    RemoveVideoFromSpaceRequest,
    Space, 
    Video,
    VideoServiceClient 
} from "../proto/videoservice"

export const $spaces = atom<Space[]>([])
export const $spaceVideos = atom<Record<string, Video[]>>({})

onMount($spaces, () => {
    console.log("spaces.ts -> onMount()")
    fetchSpaces()
})

const unaryInterceptor: UnaryInterceptor<any, any> = {
    intercept: (request, invoker) => {
        const m = request.getMetadata();
        const token = $authToken.get();
        m["authorization"] = token;
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

export const fetchSpaces = async () => {
    try {
        const response = await videoService.ListSpaces(ListSpacesRequest.fromObject({}), {})
        $spaces.set(response.spaces)
    } catch (error) {
        console.error("Error fetching spaces:", error)
    }
}

export const createSpace = async (name: string, description: string = "") => {
    try {
        const response = await videoService.CreateSpace(CreateSpaceRequest.fromObject({
            name,
            description
        }), {})
        
        // Add the new space to the list
        const currentSpaces = $spaces.get()
        $spaces.set([response, ...currentSpaces])
        
        return response
    } catch (error) {
        console.error("Error creating space:", error)
        throw error
    }
}

export const fetchSpace = async (spaceId: string) => {
    try {
        const response = await videoService.GetSpace(GetSpaceRequest.fromObject({
            space_id: spaceId
        }), {})
        return response
    } catch (error) {
        console.error("Error fetching space:", error)
        throw error
    }
}

export const fetchVideosInSpace = async (spaceId: string) => {
    try {
        const response = await videoService.ListVideosInSpace(ListVideosInSpaceRequest.fromObject({
            space_id: spaceId
        }), {})
        
        // Update the space videos cache
        const currentSpaceVideos = $spaceVideos.get()
        $spaceVideos.set({
            ...currentSpaceVideos,
            [spaceId]: response.videos
        })
        
        return response.videos
    } catch (error) {
        console.error("Error fetching videos in space:", error)
        throw error
    }
}

export const addVideoToSpace = async (videoId: string, spaceId: string) => {
    try {
        await videoService.AddVideoToSpace(AddVideoToSpaceRequest.fromObject({
            video_id: videoId,
            space_id: spaceId
        }), {})
        
        // Refresh the videos in space
        await fetchVideosInSpace(spaceId)
    } catch (error) {
        console.error("Error adding video to space:", error)
        throw error
    }
}

export const removeVideoFromSpace = async (videoId: string, spaceId: string) => {
    try {
        await videoService.RemoveVideoFromSpace(RemoveVideoFromSpaceRequest.fromObject({
            video_id: videoId,
            space_id: spaceId
        }), {})
        
        // Refresh the videos in space
        await fetchVideosInSpace(spaceId)
    } catch (error) {
        console.error("Error removing video from space:", error)
        throw error
    }
}

// Helper function to refresh space videos (useful for external calls)
export const refreshSpaceVideos = async (spaceId: string) => {
    return await fetchVideosInSpace(spaceId)
} 