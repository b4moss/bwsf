export const faqListInjectionKey = "doc-site-faq-list";

export type FaqListContext = {
  registerPanel: (id: string) => void;
  unregisterPanel: (id: string) => void;
  isOpen: (id: string) => boolean;
  toggle: (id: string) => void;
};
