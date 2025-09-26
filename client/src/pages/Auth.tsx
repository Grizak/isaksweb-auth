export default function Auth({ state }: { state: "login" | "register" }) {
  return (
    <>
      {state === "login" && <h2>Sign In</h2>}
      {state === "register" && <h2>Sign Up</h2>}
    </>
  );
}
