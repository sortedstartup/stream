import { useStore } from '@nanostores/react' // or '@nanostores/preact'
import { $videos } from '../stores/videos'


const ListOfVideos = () => {

    const videos = useStore($videos)

    return (
        <div>
            <h1>List of Videos</h1>
            <ul>
                {videos.map((video) => (
                    <li key={video.id}>{video.title}</li>
                ))}
            </ul>
        </div>
    )
}

export default ListOfVideos