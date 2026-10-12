import { Component, inject } from '@angular/core';
import { FormsModule, NgForm } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { finalize } from 'rxjs';
import { AuthService } from '../auth.service';
import { errorMessage } from '../error-message';
import { Registration } from '../models';

@Component({
  selector: 'app-register',
  imports: [FormsModule, RouterLink],
  templateUrl: './register.html',
})
export class RegisterPage {
  private auth = inject(AuthService);
  private router = inject(Router);
  data: Registration = { login: '', password: '', fullName: '', phone: '', email: '' };
  readonly phonePattern = String.raw`8\([0-9]{3}\)[0-9]{3}-[0-9]{2}-[0-9]{2}`;
  busy = false;
  error = '';

  submit(form: NgForm) {
    if (this.busy) return;
    if (form.invalid) {
      this.error = 'Проверьте обязательные поля и форматы данных.';
      return;
    }
    this.error = '';
    this.busy = true;
    const data = { ...this.data, login: this.data.login.trim(), fullName: this.data.fullName.trim() };
    this.auth.register(data).pipe(finalize(() => this.busy = false)).subscribe({
      next: () => void this.router.navigate(['/requests']),
      error: error => this.error = errorMessage(error),
    });
  }
}
