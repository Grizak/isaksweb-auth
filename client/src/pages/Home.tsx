import { Link } from "react-router";

export default function Home() {
  return (
    <>
      <aside>
        {/* Side bar */}
        <ul>
          <li>
            <Link to="/">Home</Link>
            <Link to="/login">Sign In</Link>
            <Link to="/register">Sign Up</Link>
          </li>
        </ul>
      </aside>
    </>
  );
}
