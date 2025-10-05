import type { SetStateAction } from "react";

export interface BackendMetadataResponse {
  title: string;
  metadata: {
    [key: string]: string;
  };
}

export interface UserContext {
  user: object;
  setUser: (value: SetStateAction<object>) => void;
}
