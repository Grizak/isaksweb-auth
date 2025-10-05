import LoginPage from "@/components/Login";
import RegisterPage from "@/components/Register";

export default function Auth({ state }: { state: "login" | "register" }) {
  return (
    <>
      {state === "login" && <LoginPage />}
      {state === "register" && <RegisterPage />}
    </>
  );
}
