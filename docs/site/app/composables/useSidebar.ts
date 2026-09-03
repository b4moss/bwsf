export function useSidebar() {
  const open = useState("docs-sidebar-open", () => false);

  function toggle() {
    open.value = !open.value;
  }

  function close() {
    open.value = false;
  }

  return { open, toggle, close };
}
