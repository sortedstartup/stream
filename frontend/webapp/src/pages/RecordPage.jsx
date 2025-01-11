import { Layout } from "../components/layout/Layout";
import ScreenRecorder from "../components/ScreenRecorder";

export const RecordPage = () => {
  return (
    <Layout>
      <div className="space-y-8">
        <div className="text-center">
                <h1 className="text-3xl font-bold mb-4">Record</h1>
                <p className="mb-4">Start recording your screen or camera</p>
                <div className="flex justify-center">   
                    <ScreenRecorder />
                </div>
        </div>
      </div>
    </Layout>
  )
} 