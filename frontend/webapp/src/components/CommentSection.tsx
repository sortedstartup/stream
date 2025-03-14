import React, { useState, useEffect } from "react";
import { useStore } from "@nanostores/react";
import { $comments, fetchComments, createComment } from "../stores/comments";
import { useParams } from "react-router";

const CommentSection: React.FC = () => {
    const { id: videoId } = useParams();
    const comments = useStore($comments);
    const [newComment, setNewComment] = useState("");
    const [replyText, setReplyText] = useState<{ [key: string]: string }>({});
    const [replyingTo, setReplyingTo] = useState<string | null>(null);

    useEffect(() => {
        if (videoId) {
            fetchComments(videoId);
        }
    }, [videoId]);

    const handleAddComment = async () => {
        if (!newComment.trim() || !videoId) return;
        await createComment(videoId, newComment);
        setNewComment("");
    };

    const handleReplyClick = (commentId: string) => {
        setReplyingTo(commentId === replyingTo ? null : commentId);
    };

    const handleReplyChange = (commentId: string, text: string) => {
        setReplyText((prev) => ({ ...prev, [commentId]: text }));
    };

    const handleAddReply = async (parentCommentId: string) => {
        if (!replyText[parentCommentId]?.trim() || !videoId) return;
        await createComment(videoId, replyText[parentCommentId], parentCommentId);
        setReplyText((prev) => ({ ...prev, [parentCommentId]: "" }));
        setReplyingTo(null);
    };

    const formatTimestamp = (timestamp?: { seconds: number }) => {
        if (!timestamp || typeof timestamp.seconds !== "number") return "Unknown";
        return new Date(timestamp.seconds * 1000).toLocaleString();
    };

    return (
        <div className="mt-6 p-4 bg-base-200 rounded-lg">
            <h2 className="text-xl font-semibold mb-4">Comments</h2>

            {/* Add Comment */}
            <div className="mb-4 flex items-center gap-2">
                <input
                    type="text"
                    className="input input-bordered w-full"
                    placeholder="Write a comment..."
                    value={newComment}
                    onChange={(e) => setNewComment(e.target.value)}
                />
                <button onClick={handleAddComment} className="btn btn-primary">
                    Post
                </button>
            </div>

            {/* List Comments */}
            <div className="space-y-4">
                {comments.map((comment) => (
                    <div key={comment.id} className="p-3 bg-base-100 rounded-lg shadow">
                        <p className="font-medium">{comment.content}</p>
                        <div className="text-sm text-gray-500">
                            By {comment.user_id} • {formatTimestamp(comment.created_at)}
                        </div>

                        {/* Reply Button */}
                        <button
                            className="text-blue-500 text-sm mt-2"
                            onClick={() => handleReplyClick(comment.id)}
                        >
                            {replyingTo === comment.id ? "Cancel" : "Reply"}
                        </button>

                        {/* Reply Input */}
                        {replyingTo === comment.id && (
                            <div className="mt-2 flex items-center gap-2">
                                <input
                                    type="text"
                                    className="input input-bordered w-full"
                                    placeholder="Write a reply..."
                                    value={replyText[comment.id] || ""}
                                    onChange={(e) => handleReplyChange(comment.id, e.target.value)}
                                />
                                <button
                                    onClick={() => handleAddReply(comment.id)}
                                    className="btn btn-secondary"
                                >
                                    Reply
                                </button>
                            </div>
                        )}

                        {/* Replies (if any) */}
                        {Array.isArray(comment.replies) && comment.replies.length > 0 && (
                            <div className="mt-3 ml-6 border-l pl-3 space-y-2">
                                {comment.replies.map((reply) => (
                                    <div key={reply.id} className="text-sm">
                                        <p>{reply.content}</p>
                                        <div className="text-xs text-gray-500">
                                            By {reply.user_id} • {formatTimestamp(reply.created_at)}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                ))}
            </div>
        </div>
    );
};

export default CommentSection;
