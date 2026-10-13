import { Component, inject, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { FormsModule } from '@angular/forms';

interface Booking {
 id?: number;
 fullName: string;
 room: string;
 startsAt: string;
}

@Component({
 selector: 'app-root',
 imports: [FormsModule],
 templateUrl: './app.html',
})
export class App implements OnInit {
 private http = inject(HttpClient);
 bookings: Booking[] = [];
 fullName = '';
 room = '';
 startsAt = '';
 editingId: number | undefined;
 message = '';

 ngOnInit(): void { this.loadBookings(); }

 loadBookings(): void {
  this.http.get<Booking[]>('/api/requests').subscribe({
   next: bookings => this.bookings = bookings,
   error: err => this.message = 'Ошибка загрузки: ' + err.status,
  });
 }

 submit(): void {
  const booking: Booking = {
   fullName: this.fullName.trim(), room: this.room.trim(), startsAt: this.startsAt,
  };
  if (!booking.fullName || !booking.room || !booking.startsAt) {
   this.message = 'Заполните все поля';
   return;
  }
  if (this.editingId !== undefined) {
   this.http.put<void>(`/api/requests/${this.editingId}`, booking).subscribe({
    next: () => {
     this.resetForm();
     this.message = 'Изменения сохранены';
     this.loadBookings();
    },
    error: err => this.message = 'Ошибка сохранения: ' + err.status,
   });
   return;
  }
  this.http.post<Booking>('/api/requests', booking).subscribe({
   next: created => {
    this.bookings.unshift(created);
    this.resetForm();
    this.message = 'Заявка создана';
   },
   error: err => this.message = 'Ошибка создания: ' + err.status,
  });
 }

 editBooking(booking: Booking): void {
  this.editingId = booking.id;
  this.fullName = booking.fullName;
  this.room = booking.room;
  this.startsAt = booking.startsAt;
  this.message = '';
 }

 deleteBooking(id: number | undefined): void {
  if (id === undefined) return;
  this.http.delete<void>(`/api/requests/${id}`).subscribe({
   next: () => {
    this.bookings = this.bookings.filter(booking => booking.id !== id);
    if (this.editingId === id) this.resetForm();
    this.message = 'Заявка удалена';
   },
   error: err => this.message = 'Ошибка удаления: ' + err.status,
  });
 }

 resetForm(): void {
  this.editingId = undefined;
  this.fullName = '';
  this.room = '';
  this.startsAt = '';
  this.message = '';
 }
}
