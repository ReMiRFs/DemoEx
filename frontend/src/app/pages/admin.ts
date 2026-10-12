import { Component, inject, OnInit } from '@angular/core';
import { DatePipe } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { finalize } from 'rxjs';
import { ApiService } from '../api.service';
import { PAYMENT_LABELS, STATUS_LABELS } from '../config';
import { errorMessage } from '../error-message';
import { Booking, BookingStatus } from '../models';

@Component({ selector: 'app-admin', imports: [FormsModule, DatePipe], templateUrl: './admin.html' })
export class AdminPage implements OnInit {
  private api = inject(ApiService);
  readonly statuses = STATUS_LABELS;
  readonly payments = PAYMENT_LABELS;
  bookings: Booking[] = [];
  selectedStatus: Record<number, BookingStatus> = {};
  filter = '';
  loading = false;
  busyId: number | null = null;
  error = '';
  message = '';

  ngOnInit() { this.load(); }

  load() {
    this.loading = true;
    this.error = '';
    this.api.adminList(this.filter).pipe(finalize(() => this.loading = false)).subscribe({
      next: bookings => {
        this.bookings = bookings;
        for (const booking of bookings) this.selectedStatus[booking.id] = booking.status;
      },
      error: error => this.error = errorMessage(error),
    });
  }

  save(booking: Booking) {
    if (this.busyId !== null) return;
    this.busyId = booking.id;
    this.error = '';
    this.message = '';
    this.api.updateStatus(booking.id, this.selectedStatus[booking.id]).pipe(finalize(() => this.busyId = null)).subscribe({
      next: updated => {
        this.bookings = this.bookings.map(item => item.id === updated.id ? updated : item)
          .filter(item => !this.filter || item.status === this.filter);
        this.message = `Статус заявки №${updated.id} сохранён.`;
      },
      error: error => {
        this.selectedStatus[booking.id] = booking.status;
        this.error = errorMessage(error);
      },
    });
  }
}
