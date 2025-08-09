import { atom, onMount } from "nanostores"
import { UnaryInterceptor } from "grpc-web";
import { $authToken } from "../auth/store/auth";
import { $currentTenant } from "./tenants";
import { 
    GetVideoRequest, 
    ListVideosRequest, 
    Video, 
    VideoServiceClient,
    MoveVideoToChannelRequest,
    RemoveVideoFromChannelRequest,
    DeleteVideoRequest,
    UpdateVideoRequest,
    Visibility
} from "../proto/videoservice"

export const $videos = atom<Video[]>([])
export const $tenantVideos = atom<Video[]>([]) // Videos not assigned to any channel

onMount($videos,() => {
    console.log("videos.ts -> onMount()")
    fetchVideos()
})

const unaryInterceptor: UnaryInterceptor<any, any> = {
    intercept: (request, invoker) => {
      const m = request.getMetadata();
      const token = $authToken.get();
      const currentTenant = $currentTenant.get();
      
      m["authorization"] = token;
      
      // Add tenant ID header if available
      if (currentTenant?.tenant?.id) {
        m["x-tenant-id"] = currentTenant.tenant.id;
      }
      
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
    try {
        const response = await videoService.ListVideos(ListVideosRequest.fromObject({
            page_number: 0,
            page_size: 10,
        }),{})

        $videos.set(response.videos)
    } catch (error) {
        console.error("Error fetching videos:", error)
        // Clear videos on error (especially auth errors)
        $videos.set([])
        throw error // Re-throw to let calling code handle if needed
    }
}

export const fetchVideo = async (id: string) => {
    try {
        const response = await videoService.GetVideo(GetVideoRequest.fromObject({
             video_id: id
        }),{})

        return response
    } catch (error) {
        console.error("Error fetching video:", error)
        throw error
    }
}

// Fetch videos that are not assigned to any channel (user's private videos)
export const fetchTenantVideos = async () => {
    try {
        // Get all accessible videos, then filter for private videos (no channel)
        const response = await videoService.ListVideos(ListVideosRequest.fromObject({
            page_number: 0,
            page_size: 100,
        }),{})

        // Filter for videos without channels (user's private videos)
        const privateVideos = response.videos.filter(video => !video.channel_id || video.channel_id === '')
        $tenantVideos.set(privateVideos)
        return privateVideos
    } catch (error) {
        console.error("Error fetching tenant videos:", error)
        $tenantVideos.set([])
        throw error
    }
}

// Video management functions
export const moveVideoToChannel = async (videoId: string, channelId: string): Promise<void> => {
    try {
        await videoService.MoveVideoToChannel(MoveVideoToChannelRequest.fromObject({
            video_id: videoId,
            channel_id: channelId
        }), {})
        
        // Refresh videos to reflect the change
        await fetchVideos()
        await fetchTenantVideos()
    } catch (error) {
        console.error("Error moving video to channel:", error)
        throw error
    }
}

export const removeVideoFromChannel = async (videoId: string): Promise<void> => {
    try {
        await videoService.RemoveVideoFromChannel(RemoveVideoFromChannelRequest.fromObject({
            video_id: videoId
        }), {})
        
        // Refresh videos to reflect the change
        await fetchVideos()
        await fetchTenantVideos()
    } catch (error) {
        console.error("Error removing video from channel:", error)
        throw error
    }
}

export const deleteVideo = async (videoId: string): Promise<void> => {
    try {
        await videoService.DeleteVideo(DeleteVideoRequest.fromObject({
            video_id: videoId
        }), {})
        
        // Refresh videos to reflect the change
        await fetchVideos()
        await fetchTenantVideos()
    } catch (error) {
        console.error("Error deleting video:", error)
        throw error
    }
}

export const updateVideo = async (
  videoId: string,
  title: string,
  description: string
): Promise<void> => {
  try {
    console.log({
        video_id: videoId,
        title,
        description,
    });

    const request = UpdateVideoRequest.fromObject({
      video_id: videoId,
      title,
      description,
    });

    await videoService.UpdateVideo(request, {});

    await fetchVideos();
    await fetchTenantVideos();
  } catch (error) {
    console.error("Error updating video:", error);
    throw error;
  }
};

