'use client';

import { ListVideosRequest, Video, VideoServiceClient } from "@/proto/videoservice"
import { atom, onMount } from "nanostores"
import { UnaryInterceptor } from "grpc-web";
import { $authToken } from "../auth/store/user";

export const $videos = atom<Video[]>([])

onMount($videos,() => {
    console.log("videos.ts -> onMount()")
    // fetchVideos()
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
    "http://127.0.0.1:8080",
    {},
    {
        unaryInterceptors: [unaryInterceptor],
    }
);

export const fetchVideos = async () => {
    const response = await videoService.ListVideos(ListVideosRequest.fromObject({
        pageNumber: 1,
        pageSize: 10,
    }),{})

    $videos.set(response.videos)
}