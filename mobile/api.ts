import { Image } from 'react-native';

import type { ReactNativeFile, Role } from './types';

export const API_URL = process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:8080';

export const BACKEND_UNREACHABLE =
  'Backend erişilemiyor. cd backend && go run .';

const PLACEHOLDER_ICON = require('./assets/icon.png');

export function getPlaceholderImage(): ReactNativeFile | null {
  const source = Image.resolveAssetSource(PLACEHOLDER_ICON);
  if (!source?.uri) {
    return null;
  }
  return {
    uri: source.uri,
    name: 'report.png',
    type: 'image/png',
  };
}

export function appendFile(form: FormData, field: string, file: ReactNativeFile): void {
  form.append(field, file as unknown as Blob);
}

export async function apiFetch(
  path: string,
  role: Role,
  init?: RequestInit,
): Promise<Response> {
  const headers = new Headers(init?.headers);
  headers.set('X-Role', role);
  return fetch(`${API_URL}${path}`, { ...init, headers });
}
