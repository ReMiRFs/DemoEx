import { inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Booking, BookingInput, BookingStatus } from './models';

@Injectable({ providedIn: 'root' })
export class ApiService {
  private http = inject(HttpClient);

  list() { return this.http.get<Booking[]>('/api/requests'); }
  create(data: BookingInput) { return this.http.post<Booking>('/api/requests', data); }
  update(id: number, data: BookingInput) {
    return this.http.put<Booking>(`/api/requests/${id}`, data);
  }
  remove(id: number) { return this.http.delete<void>(`/api/requests/${id}`); }
  review(id: number, review: string) {
    return this.http.post<Booking>(`/api/requests/${id}/review`, { review });
  }
  adminList(status: string) {
    return this.http.get<Booking[]>('/api/admin/requests', { params: { status } });
  }
  updateStatus(id: number, status: BookingStatus) {
    return this.http.patch<Booking>(`/api/admin/requests/${id}/status`, { status });
  }
}
