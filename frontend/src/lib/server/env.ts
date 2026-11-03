import { env } from "$env/dynamic/private";

export const API_URL = env.API_URL ?? "http://localhost:8080";
