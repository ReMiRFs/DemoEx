import { Component, inject } from '@angular/core';
import { FormsModule, NgForm } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { finalize } from 'rxjs';
import { AuthService } from '../auth.service';
import { errorMessage } from '../error-message';

@Component({ selector: 'app-login', imports: [FormsModule, RouterLink], templateUrl: './login.html' })
export class LoginPage {
  private auth = inject(AuthService);
  private router = inject(Router);
  login = '';
  password = '';
  busy = false;
  error = '';

  submit(form: NgForm) {
    if (this.busy) return;
    if (form.invalid || !this.login.trim()) {
      this.error = 'Введите логин и пароль.';
      return;
    }
    this.busy = true;
    this.error = '';
    this.auth.login(this.login.trim(), this.password).pipe(finalize(() => this.busy = false)).subscribe({
      next: user => void this.router.navigate([user.role === 'admin' ? '/admin' : '/requests']),
      error: error => this.error = errorMessage(error),
    });
  }
}
