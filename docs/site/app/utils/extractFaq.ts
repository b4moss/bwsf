export type FaqQa = {
  id: string;
  question: string;
  answer: string;
};

/** Flatten minimark / MDC node trees to plain text. */
export function minimarkToText(input: unknown): string {
  if (input == null || typeof input === "boolean") {
    return "";
  }
  if (typeof input === "string" || typeof input === "number") {
    return String(input);
  }
  if (Array.isArray(input)) {
    // Minimark element: [tag, props, ...children]
    if (typeof input[0] === "string") {
      return input.slice(2).map(minimarkToText).join("");
    }
    return input.map(minimarkToText).join("");
  }
  if (typeof input === "object") {
    const record = input as Record<string, unknown>;
    if ("value" in record) {
      return minimarkToText(record.value);
    }
    if ("children" in record) {
      return minimarkToText(record.children);
    }
  }
  return "";
}

/**
 * Collect FAQ Q/A pairs from a Nuxt Content page body (minimark). This is the
 * only source of FAQPage JSON-LD, so the output does not depend on component
 * render order.
 */
export function extractFaqFromBody(body: unknown): FaqQa[] {
  const results: FaqQa[] = [];
  let index = 0;

  function walk(node: unknown): void {
    if (!Array.isArray(node)) {
      if (node && typeof node === "object" && "value" in node) {
        walk((node as { value: unknown }).value);
      }
      return;
    }

    if (typeof node[0] === "string") {
      const tag = node[0];
      const props = (node[1] || {}) as Record<string, unknown>;
      const children = node.slice(2);

      if (tag === "faq-item") {
        const question = String(props.question || "").trim();
        const answer = minimarkToText(children).replace(/\s+/g, " ").trim();
        if (question && answer) {
          results.push({
            id: `body-faq-${index++}`,
            question,
            answer,
          });
        }
      }

      for (const child of children) {
        walk(child);
      }
      return;
    }

    for (const child of node) {
      walk(child);
    }
  }

  walk(body);
  return results;
}
