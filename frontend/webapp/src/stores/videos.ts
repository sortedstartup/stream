import { atom, onMount } from "nanostores"
import { UnaryInterceptor } from "grpc-web";
import { $authToken } from "../auth/store/auth";
import { $currentTenant } from "./tenants";
import { GetVideoRequest, ListVideosRequest, Video, VideoServiceClient } from "../proto/videoservice"

export const $videos = atom<Video[]>([])

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
            pageNumber: 0,
            pageSize: 10,
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