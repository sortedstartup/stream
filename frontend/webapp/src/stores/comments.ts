import { atom, onMount } from "nanostores";
import { 
  CommentServiceClient, 
  CreateCommentRequest, 
  ListCommentsRequest, 
  UpdateCommentRequest, 
  DeleteCommentRequest, 
  Comment 
} from "../proto/commentservice";
import { $authToken } from "../auth/store/auth";

interface CommentWithReplies extends Comment {
  replies: Comment[];
}

// Store for comments
export const $comments = atom<CommentWithReplies[]>([]);

export const commentService = new CommentServiceClient(
  import.meta.env.VITE_PUBLIC_API_URL,
  {},
  {
    unaryInterceptors: [
      {
        intercept: (request, invoker) => {
          const metadata = request.getMetadata();
          metadata["authorization"] = $authToken.get();
          return invoker(request);
        },
      },
    ],
  }
);

export const fetchComments = async (videoId: string) => {
  try {
      console.log("Fetching comments for video:", videoId);

      const request = new ListCommentsRequest();
      request.video_id = videoId;

      const response = await commentService.ListComments(request, {});
      
      console.log("Comments fetched successfully:", response);
      $comments.set(response.comments as CommentWithReplies[]);
  } catch (error: unknown) {
      if (error instanceof SyntaxError) {
          console.error("Invalid JSON response. Possible server error.");
      } else if (error instanceof Error) {
          console.error("Error fetching comments:", error.message);
      } else {
          console.error("Unknown error fetching comments:", error);
      }
  }
};

export const createComment = async (
  videoId: string,
  content: string,
  parentCommentId?: string
) => {
  try {
      console.log("Creating comment with data:", { videoId, content, parentCommentId });

      const request = new CreateCommentRequest();
      request.video_id = videoId;
      request.content = content;
      if (parentCommentId) request.parent_comment_id = parentCommentId;

      const response = await commentService.CreateComment(request, {});
      console.log("Comment created successfully:", response);

      // Refresh comments after adding a new one
      await fetchComments(videoId);
  } catch (error: unknown) {
      if (error instanceof Error) {
          console.error("Error creating comment:", error.message);
      } else {
          console.error("Error creating comment:", error);
      }
  }
};

export const updateComment = async (commentId: string, content: string) => {
  try {
    const request = new UpdateCommentRequest();
    request.comment_id = commentId;
    request.content = content;

    await commentService.UpdateComment(request, null); 

    const updatedComments = $comments.get();
    const commentIndex = updatedComments.findIndex(comment => comment.id === commentId);
    
    if (commentIndex !== -1) {
      updatedComments[commentIndex].content = content;
      $comments.set(updatedComments);
    }
  } catch (error) {
    console.error("Failed to update comment:", error);
  }
};

// Delete a comment
export const deleteComment = async (commentId: string) => {
  try {
    const request = new DeleteCommentRequest();
    request.comment_id = commentId;

    await commentService.DeleteComment(request, null); 

    // Remove the deleted comment from the store
    $comments.set($comments.get().filter(comment => comment.id !== commentId));
  } catch (error) {
    console.error("Failed to delete comment:", error);
  }
};
