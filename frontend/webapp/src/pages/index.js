import Image from "next/image";
import localFont from "next/font/local";

const geistSans = localFont({
  src: "./fonts/GeistVF.woff",
  variable: "--font-geist-sans",
  weight: "100 900",
});
const geistMono = localFont({
  src: "./fonts/GeistMonoVF.woff",
  variable: "--font-geist-mono",
  weight: "100 900",
});

export default function Home() {
  return (
    <div
      className={`${geistSans.variable} ${geistMono.variable} grid grid-rows-[20px_1fr_20px] items-center justify-items-center min-h-screen p-8 pb-20 gap-16 sm:p-20 font-[family-name:var(--font-geist-sans)]`}
    >
      {/* Header */}
      <div className="text-lg font-bold">
        <h1>Screen Recorder App</h1>
      </div>

      {/* Main Content */}
      <main className="flex flex-col gap-8 row-start-2 items-center">
        <button
          className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-all"
          onClick={() => {
            // You can navigate to /screen-recorder or trigger functionality here
            window.location.href = "/screen-recorder";
          }}
        >
          + Record Video
        </button>
      </main>

    </div>
  );
}
