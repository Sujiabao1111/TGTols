import { NextResponse } from "next/server"
import type { NextRequest } from "next/server"

export function proxy(request: NextRequest) {
  // Check if accessing protected route
  const isProtectedRoute =
    request.nextUrl.pathname.startsWith("/wallet") ||
    request.nextUrl.pathname.startsWith("/profile") ||
    request.nextUrl.pathname.startsWith("/invite") ||
    request.nextUrl.pathname.startsWith("/activity") ||
    request.nextUrl.pathname.startsWith("/favorites") ||
    request.nextUrl.pathname.startsWith("/game")

  // Client-side auth check is handled in the protected layout
  // This middleware just allows the request to proceed
  if (isProtectedRoute) {
    return NextResponse.next()
  }

  return NextResponse.next()
}

export const config = {
  matcher: ["/wallet/:path*", "/profile/:path*", "/invite/:path*", "/activity/:path*", "/favorites/:path*", "/game/:path*"],
}
