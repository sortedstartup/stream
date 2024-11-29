import Link from "next/link";

export default function Home() {
  return (
    <div style={{ textAlign: "center", padding: "20px" }}>
      <h1>Welcome to the Home Page</h1>
      <Link
        href="/screen-recorder"
        style={{
          display: "inline-block",
          marginTop: "20px",
          padding: "10px 20px",
          background: "#0070f3",
          color: "#fff",
          textDecoration: "none",
          borderRadius: "5px",
        }}
      >
        Go to Screen Recorder
      </Link>
    </div>
  );
}

