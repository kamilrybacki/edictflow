// Handle tabbed content switching
(function() {
  window.initTabs = function initTabs() {
    document.querySelectorAll('.tabbed-set').forEach(function(tabSet) {
      var inputs = tabSet.querySelectorAll(':scope > input[type="radio"]');
      var labelsContainer = tabSet.querySelector('.tabbed-labels');
      var contentContainer = tabSet.querySelector('.tabbed-content');

      if (!labelsContainer || !contentContainer) return;

      var labels = labelsContainer.querySelectorAll('label');
      var blocks = contentContainer.querySelectorAll('.tabbed-block');

      function showTab(index) {
        // Update labels
        labels.forEach(function(label, i) {
          if (i === index) {
            label.classList.add('tabbed-active');
          } else {
            label.classList.remove('tabbed-active');
          }
        });

        // Update content blocks
        blocks.forEach(function(block, i) {
          if (i === index) {
            block.style.cssText = 'display: block !important; visibility: visible !important;';
          } else {
            block.style.cssText = 'display: none !important;';
          }
        });

        // Update radio input
        if (inputs[index]) {
          inputs[index].checked = true;
        }
      }

      // Add click handlers to labels
      labels.forEach(function(label, index) {
        label.addEventListener('click', function(e) {
          e.preventDefault();
          showTab(index);
        });
      });

      // Initialize: show the checked tab or first tab
      var checkedIndex = 0;
      inputs.forEach(function(input, i) {
        if (input.checked) {
          checkedIndex = i;
        }
      });
      showTab(checkedIndex);
    });
  }

  // Run on DOM ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initTabs);
  } else {
    initTabs();
  }

  // Also run after a short delay in case of dynamic content
  setTimeout(initTabs, 100);
})();
