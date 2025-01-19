import { useStore } from '@nanostores/react'
import { $videos } from '../stores/videos'
import { VideoStatus, Visibility } from '../proto/videoservice'
import { useNavigate } from 'react-router'

const VideoCard = ({ video }) => {
    const navigate = useNavigate()

    const getStatusBadge = (status) => {
        const statusClasses = {
            [VideoStatus.STATUS_PROCESSING]: "badge badge-warning",
            [VideoStatus.STATUS_READY]: "badge badge-success",
            [VideoStatus.STATUS_FAILED]: "badge badge-error",
            [VideoStatus.STATUS_UNSPECIFIED]: "badge badge-ghost",
        }
        
        const statusText = VideoStatus[status].replace('STATUS_', '')
        return (
            <span className={statusClasses[status]}>
                {statusText}
            </span>
        )
    }

    return (
        <div 
            className="card bg-base-100 shadow-xl hover:shadow-2xl transition-shadow duration-300 cursor-pointer" 
            onClick={() => navigate(`/video/${video.id}`)}
        >
            {/* Thumbnail */}
            <figure className="relative aspect-video">
                {video.thumbnail_url ? (
                    <img 
                        src={video.thumbnail_url} 
                        alt={video.title}
                        className="w-full h-full object-cover"
                    />
                ) : (
                    <div className="absolute inset-0 flex items-center justify-center">
                        <svg className="w-12 h-12 text-base-content opacity-40" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                        </svg>
                    </div>
                )}
                {/* Status Badge */}
                <div className="absolute top-2 right-2">
                    {getStatusBadge(video.status)}
                </div>
            </figure>

            {/* Content */}
            <div className="card-body p-4">
                <h3 className="card-title text-base-content truncate">{video.title}</h3>
                <p className="text-sm text-base-content/70 line-clamp-2">{video.description}</p>
                
                {/* Footer */}
                <div className="flex items-center justify-between mt-4">
                    <div className="flex items-center gap-2">
                        <span className="text-xs text-base-content/60">
                            {new Date(video.created_at?.seconds * 1000).toLocaleDateString()}
                        </span>
                        {video.visibility !== Visibility.VISIBILITY_PRIVATE && (
                            <span className="badge badge-primary">
                                {video.visibility === Visibility.VISIBILITY_SHARED ? 'Shared' : 'Public'}
                            </span>
                        )}
                    </div>
                </div>
            </div>
        </div>
    )
}

const ListOfVideos = () => {
    const videos = useStore($videos)

    return (
        <div className="container mx-auto px-4 py-8">
            <h1 className="text-2xl font-bold mb-6">List of Videos</h1>
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-3">
                {videos.map((video) => (
                    <div key={video.id} className="max-w-[225px] mx-auto w-full">
                        <VideoCard video={video} />
                    </div>
                ))}
            </div>
        </div>
    )
}

export default ListOfVideos