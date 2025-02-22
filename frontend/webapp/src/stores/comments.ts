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
export const $comments = atom<Comment[]>([]);

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

// Fetch comments for a specific video
export const fetchComments = async (videoId: string) => {
  try {
      const response = await commentService.ListComments(
          new ListCommentsRequest({ page_size: 50, page_number: 0, video_id: videoId }),
          {}
      );

      const commentsMap: Record<string, CommentWithReplies> = {};
      const rootComments: CommentWithReplies[] = [];

      response.comments.forEach(comment => {
        commentsMap[comment.id] = Object.assign(Object.create(Object.getPrototypeOf(comment)), comment, { replies: [] });
      });

      response.comments.forEach(comment => {
          if (comment.parent_comment_id && commentsMap[comment.parent_comment_id]) {
              commentsMap[comment.parent_comment_id].replies.push(commentsMap[comment.id]);
          } else {
              rootComments.push(commentsMap[comment.id]);
          }
      });

      const sortReplies = (comment: CommentWithReplies) => {
          comment.replies.sort((a, b) => (a.created_at?.seconds || 0) - (b.created_at?.seconds || 0));
          comment.replies.forEach(sortReplies);
      };

      rootComments.forEach(sortReplies);

      $comments.set(rootComments);
  } catch (error) {
      console.error("Failed to fetch comments:", error);
  }
};

// Create a new comment
export const createComment = async (videoId: string, content: string, parentCommentId?: string) => {
    try {
        const request = CreateCommentRequest.fromObject({
            content,
            video_id: videoId,
            parent_comment_id: parentCommentId || undefined,
        });

        const response = await commentService.CreateComment(request, {});
        fetchComments(videoId); // Refresh comments after adding a new one
    } catch (error) {
        console.error("Failed to create comment:", error);
    }
};

// Update a comment
export const updateComment = async (commentId: string, content: string) => {
  try {
    const request = UpdateCommentRequest.fromObject({
      comment_id: commentId,
      content,
    });

    const updatedComment = await commentService.UpdateComment(request, {});

    $comments.set(
      $comments.get().map((comment) => 
        comment.id === updatedComment.id ? updatedComment : comment
      )
    );
  } catch (error) {
    console.error("Error updating comment:", error);
  }
};

// Delete a comment
export const deleteComment = async (commentId: string) => {
  try {
    const request = DeleteCommentRequest.fromObject({
      comment_id: commentId,
    });

    await commentService.DeleteComment(request, {});

    $comments.set($comments.get().filter((comment) => comment.id !== commentId));
  } catch (error) {
    console.error("Error deleting comment:", error);
  }
};

// Auto-fetch comments when store is mounted
onMount($comments, () => {
  console.log("comments.ts -> onMount()");
});
