import { Component, inject } from '@angular/core';
import { Title } from '@angular/platform-browser';
import { Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { finalize } from 'rxjs';
import { AuthService } from './auth.service';
import { APP } from './config';
import { errorMessage } from './error-message';

@Component({
  selector: 'app-root',
  imports: [RouterOutlet, RouterLink, RouterLinkActive],
  templateUrl: './app.html',
})
export class App {
  readonly app = APP;
  readonly auth = inject(AuthService);
  private router = inject(Router);
  error = '';
  busy = false;

  constructor() { inject(Title).setTitle(APP.title); }

  logout() {
    if (this.busy) return;
    this.busy = true;
    this.error = '';
    this.auth.logout().pipe(finalize(() => this.busy = false)).subscribe({
      next: () => void this.router.navigate(['/login']),
      error: error => this.error = errorMessage(error),
    });
  }
}
