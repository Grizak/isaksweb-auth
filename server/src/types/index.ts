import { Request } from "express";

export interface RegisterRequest extends Request {
  body: {
    username: string;
    age: number;
    email: string;
    password: string;
  };
}

export interface RouteDataRequest extends Request {
  query: {
    pathname: string;
  };
}

export interface RouteData {
  [key: string]: {
    title: string;
    metadata: {
      [key: string]: string;
    };
  };
}
