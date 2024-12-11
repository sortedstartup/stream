/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  async rewrites() {
    if (process.env.NODE_ENV === "development") {
      return [
        {
          source: "/hack/:path*", // Matches `/hack` and all sub-paths
          destination: "http://localhost:8080/:path*", // Proxy to backend
        },
      ];
    }
    return [];
  },
};

export default nextConfig;
