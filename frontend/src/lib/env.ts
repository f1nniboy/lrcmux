import { env } from "$env/dynamic/public";

export const API_URL = env.PUBLIC_API_URL ?? "";
