import { HttpErrorResponse } from '@angular/common/http';

// Сервер возвращает ошибки в одном формате: { "error": "Описание" }.
export function errorMessage(error: unknown): string {
  if (error instanceof HttpErrorResponse) {
    if (error.status === 0) {
      return 'Нет соединения с сервером. Проверьте, что Go запущен.';
    }
    if (typeof error.error?.error === 'string') return error.error.error;
    if (error.status === 401) return 'Войдите в аккаунт, чтобы продолжить.';
    return `Не удалось выполнить запрос (код ${error.status}). Попробуйте ещё раз.`;
  }
  return 'Произошла ошибка. Попробуйте ещё раз.';
}
