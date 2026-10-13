import { Component, inject, OnInit } from '@angular/core';
import { FormsModule, NgForm } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { finalize } from 'rxjs';
import { ApiService } from '../api.service';
import { APP, PAYMENT_LABELS } from '../config';
import { errorMessage } from '../error-message';
import { BookingInput } from '../models';

@Component({ selector: 'app-request-form', imports: [FormsModule, RouterLink], templateUrl: './request-form.html' })
export class RequestFormPage implements OnInit {
  private api = inject(ApiService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);
  readonly rooms = APP.rooms;
  readonly payments = PAYMENT_LABELS;
  data: BookingInput = { room: '', startsAt: '', payment: 'cash' };
  id: number | null = null;
  loading = false;
  unavailable = false;
  busy = false;
  error = '';

  ngOnInit() {
    const id = this.route.snapshot.paramMap.get('id');
    if (!id) return;
    this.id = Number(id);
    this.loading = true;
    this.api.list().pipe(finalize(() => this.loading = false)).subscribe({
      next: bookings => {
        const booking = bookings.find(item => item.id === this.id);
        if (!booking) {
          this.unavailable = true;
          this.error = 'Заявка не найдена или недоступна.';
          return;
        }
        this.data = { room: booking.room, startsAt: booking.startsAt, payment: booking.payment };
      },
      error: error => { this.error = errorMessage(error); this.unavailable = true; },
    });
  }

  submit(form: NgForm) {
    if (this.busy || this.loading || this.unavailable) return;
    if (form.invalid || !this.data.room.trim()) {
      this.error = 'Укажите помещение, дату и способ оплаты.';
      return;
    }
    this.error = '';
    this.busy = true;
    const data = { ...this.data, room: this.data.room.trim() };
    const request = this.id === null ? this.api.create(data) : this.api.update(this.id, data);
    request.pipe(finalize(() => this.busy = false)).subscribe({
      next: () => void this.router.navigate(['/requests']),
      error: error => this.error = errorMessage(error),
    });
  }
}
