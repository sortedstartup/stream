import { useStore } from '@nanostores/react' // or '@nanostores/preact'
import { $videos } from '../stores/videos'


const ListOfVideos = () => {

    const videos = useStore($videos)

    return (
        <div>
            <h1>List of Videos</h1>
            {videos.map((video) => (
                <div key={video.id}>{video.title}</div>
            ))}
        </div>
    )
}

export default ListOfVideos