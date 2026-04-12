import { create } from "zustand";

type UIState = {
  isSidebarOpen: boolean;
  isRightPanelOpen: boolean;
  toggleSidebar: () => void;
  toggleRightPanel: () => void;
};

export const useUIStore = create<UIState>((set) => ({
  isSidebarOpen: true,
  isRightPanelOpen: true,
  toggleSidebar: () => set((s) => ({ isSidebarOpen: !s.isSidebarOpen })),
  toggleRightPanel: () => set((s) => ({ isRightPanelOpen: !s.isRightPanelOpen })),
}));
