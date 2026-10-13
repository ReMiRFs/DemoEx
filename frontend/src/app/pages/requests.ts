import { Component, inject, OnInit } from '@angular/core';
import { DatePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { finalize } from 'rxjs';
import { ApiService } from '../api.service';
import { AuthService } from '../auth.service';
import { PAYMENT_LABELS, STATUS_LABELS } from '../config';
import { errorMessage } from '../error-message';
import { Booking } from '../models';

@Component({ selector: 'app-requests', imports: [DatePipe, FormsModule, RouterLink], templateUrl: './requests.html' })
export class RequestsPage implements OnInit {
  private api = inject(ApiService);
  readonly auth = inject(AuthService);
  readonly statuses = STATUS_LABELS;
  readonly payments = PAYMENT_LABELS;
  bookings: Booking[] = [];
  reviews: Record<number, string> = {};
  loading = false;
  busyId: number | null = null;
  error = '';
  message = '';

  ngOnInit() { this.load(); }

  load() {
    this.loading = true;
    this.error = '';
    this.api.list().pipe(finalize(() => this.loading = false)).subscribe({
      next: bookings => {
        this.bookings = bookings;
        for (const booking of bookings) this.reviews[booking.id] = booking.review;
      },
      error: error => this.error = errorMessage(error),
    });
  }

  remove(booking: Booking) {
    if (this.busyId !== null || !window.confirm(`Удалить заявку №${booking.id}?`)) return;
    this.busyId = booking.id;
    this.error = '';
    this.message = '';
    this.api.remove(booking.id).pipe(finalize(() => this.busyId = null)).subscribe({
      next: () => {
        this.bookings = this.bookings.filter(item => item.id !== booking.id);
        this.message = 'Заявка удалена.';
      },
      error: error => this.error = errorMessage(error),
    });
  }

  saveReview(booking: Booking) {
    if (this.busyId !== null) return;
    const review = (this.reviews[booking.id] ?? '').trim();
    if (!review) { this.error = 'Введите текст отзыва.'; return; }
    this.busyId = booking.id;
    this.error = '';
    this.message = '';
    this.api.review(booking.id, review).pipe(finalize(() => this.busyId = null)).subscribe({
      next: updated => {
        this.bookings = this.bookings.map(item => item.id === updated.id ? updated : item);
        this.message = 'Спасибо! Отзыв сохранён.';
      },
      error: error => this.error = errorMessage(error),
    });
  }
}
