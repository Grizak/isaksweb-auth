export interface Page {
  url: string;
  linkTitle: string;
  isFirstElement?: boolean;
  element: React.ReactNode;
}
