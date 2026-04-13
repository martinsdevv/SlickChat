import { create } from "zustand";

type UIState = {
  isSidebarOpen: boolean;
  isRightPanelOpen: boolean;
  mobileView: "rooms" | "chat";
  toggleSidebar: () => void;
  toggleRightPanel: () => void;
  setMobileView: (view: "rooms" | "chat") => void;
};

export const useUIStore = create<UIState>((set) => ({
  isSidebarOpen: true,
  isRightPanelOpen: false,
  mobileView: "rooms",
  toggleSidebar: () => set((s) => ({ isSidebarOpen: !s.isSidebarOpen })),
  toggleRightPanel: () => set((s) => ({ isRightPanelOpen: !s.isRightPanelOpen })),
  setMobileView: (view) => set({ mobileView: view }),
}));
