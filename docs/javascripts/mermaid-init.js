// Initialize Mermaid diagrams
document.addEventListener('DOMContentLoaded', function() {
  if (typeof mermaid !== 'undefined') {
    mermaid.initialize({
      startOnLoad: true,
      theme: 'dark',
      securityLevel: 'loose',
      fontFamily: 'inherit'
    });

    // Re-render any mermaid blocks that might have been missed
    mermaid.run({
      querySelector: '.mermaid'
    });
  }
});
