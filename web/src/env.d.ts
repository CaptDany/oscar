/// <reference path="../.astro/types.d.ts" />

declare namespace App {
  interface Locals {
    user?: {
      id: string;
      email: string;
      first_name: string;
      last_name: string;
      roles: string[];
      avatar_key?: string;
      avatar_url?: string;
    };
    token?: string;
  }
}