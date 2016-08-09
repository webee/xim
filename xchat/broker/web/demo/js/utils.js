
export function decode_ns_user(user) {
  var parts = user.split(':')
  if (parts.length > 1) {
    return {ns:parts[0], user:parts[1], full_user:user}
  }
  return {ns:'', user:user, full_user:user}
}
