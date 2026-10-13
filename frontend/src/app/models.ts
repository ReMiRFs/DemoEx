export interface User {
  id: number;
  login: string;
  fullName: string;
  role: 'user' | 'admin';
}

export interface Registration {
  login: string;
  password: string;
  fullName: string;
  phone: string;
  email: string;
}

export type BookingStatus = 'new' | 'scheduled' | 'done';
export type Payment = 'cash' | 'sbp';

export interface BookingInput {
  room: string;
  startsAt: string;
  payment: Payment;
}

export interface Booking extends BookingInput {
  id: number;
  userId: number;
  fullName: string;
  status: BookingStatus;
  review: string;
}
