import { create } from "zustand";
import { persist } from "zustand/middleware";
import { apiRequest } from "../../shared/api/http-client";
import type { CreateRoomRequest, Room, RoomMember } from "../../shared/api/types";

type RoomFilters = "ALL" | "TEMPORARY" | "ZERO_LOGGING";

type RoomsState = {
  rooms: Room[];
  activeRoomId: string | null;
  membersByRoom: Record<string, RoomMember[]>;
  isLoadingRooms: boolean;
  isCreatingRoom: boolean;
  roomFilter: RoomFilters;
  setRoomFilter: (value: RoomFilters) => void;
  setActiveRoom: (roomId: string) => void;
  loadRooms: (token: string) => Promise<void>;
  createRoom: (token: string, input: CreateRoomRequest) => Promise<Room>;
  joinRoom: (token: string, roomId: string) => Promise<void>;
  addMember: (token: string, roomId: string, handle: string) => Promise<void>;
  loadMembers: (token: string, roomId: string) => Promise<void>;
  reset: () => void;
};

function sortRoomsByPriority(rooms: Room[]): Room[] {
  return [...rooms].sort((a, b) => {
    const aScore = a.zero_logging ? 3 : a.type === "TEMPORARY" ? 2 : 1;
    const bScore = b.zero_logging ? 3 : b.type === "TEMPORARY" ? 2 : 1;
    if (aScore !== bScore) {
      return bScore - aScore;
    }
    return a.name.localeCompare(b.name);
  });
}

export const useRoomsStore = create<RoomsState>()(
  persist(
    (set) => ({
      rooms: [],
      activeRoomId: null,
      membersByRoom: {},
      isLoadingRooms: false,
      isCreatingRoom: false,
      roomFilter: "ALL",
      setRoomFilter: (value) => set({ roomFilter: value }),
      setActiveRoom: (roomId) => set({ activeRoomId: roomId }),
      loadRooms: async (token) => {
        set({ isLoadingRooms: true });
        try {
          const rooms = await apiRequest<Room[]>("/rooms", { token });
          const sorted = sortRoomsByPriority(rooms);
          const activeStillExists = (activeRoomId: string | null) =>
            activeRoomId !== null && sorted.some((room) => room.room_id === activeRoomId);
          set((state) => ({
            rooms: sorted,
            activeRoomId: activeStillExists(state.activeRoomId)
              ? state.activeRoomId
              : sorted[0]?.room_id ?? null,
          }));
        } finally {
          set({ isLoadingRooms: false });
        }
      },
      createRoom: async (token, input) => {
        set({ isCreatingRoom: true });
        try {
          const room = await apiRequest<Room>("/rooms", {
            method: "POST",
            token,
            body: input,
          });

          set((state) => ({
            rooms: sortRoomsByPriority([room, ...state.rooms]),
            activeRoomId: room.room_id,
          }));

          return room;
        } finally {
          set({ isCreatingRoom: false });
        }
      },
      joinRoom: async (token, roomId) => {
        await apiRequest<void>("/room-members", {
          method: "POST",
          token,
          query: { room_id: roomId },
        });
      },
      addMember: async (token, roomId, handle) => {
        await apiRequest<void>("/room-members/add", {
          method: "POST",
          token,
          body: {
            room_id: roomId,
            handle,
          },
        });
      },
      loadMembers: async (token, roomId) => {
        const members = await apiRequest<RoomMember[]>("/room-members", {
          token,
          query: { room_id: roomId },
        });

        set((state) => ({
          membersByRoom: {
            ...state.membersByRoom,
            [roomId]: members,
          },
        }));
      },
      reset: () =>
        set({
          rooms: [],
          activeRoomId: null,
          membersByRoom: {},
          isLoadingRooms: false,
          isCreatingRoom: false,
          roomFilter: "ALL",
        }),
    }),
    {
      name: "slickchat-rooms",
      partialize: (state) => ({
        activeRoomId: state.activeRoomId,
        roomFilter: state.roomFilter,
      }),
    },
  ),
);

export function filterRooms(rooms: Room[], filter: RoomFilters): Room[] {
  if (filter === "ALL") {
    return rooms;
  }
  if (filter === "TEMPORARY") {
    return rooms.filter((room) => room.type === "TEMPORARY");
  }
  return rooms.filter((room) => room.zero_logging);
}
