/* eslint-disable react-refresh/only-export-components */
import type { UserContext as UserContextInterface } from "@/types";
import { useState, createContext, useContext, type JSX } from "react";

export const UserContext = createContext<UserContextInterface>({
  user: {},
  setUser: () => {},
});

export const UserProvider = ({
  children,
}: {
  children: JSX.Element;
}): JSX.Element => {
  const [user, setUser] = useState<UserContextInterface["user"]>({});

  return (
    <UserContext.Provider value={{ user, setUser }}>
      {children}
    </UserContext.Provider>
  );
};

const useUser = () => {
  const context = useContext(UserContext);
  if (!context) {
    throw new Error("useUser can only be used inside a UserProvider component");
  }
  return context;
};

export default useUser;
