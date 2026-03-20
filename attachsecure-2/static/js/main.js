// AttachSecure · main.js

// Streak dots: generate 7 days via JS if template helper unavailable
(function() {
  var dotsWrap = document.querySelector('.streak-dots');
  if (dotsWrap && dotsWrap.children.length === 0) {
    for (var i = 0; i < 7; i++) {
      var dot = document.createElement('div');
      dot.className = 'streak-dot' + (i < 6 ? ' filled' : '');
      dot.title = 'Day ' + (i + 1);
      dotsWrap.appendChild(dot);
    }
  }
})();

// Smooth nav active state
(function() {
  var path = window.location.pathname;
  document.querySelectorAll('.nav-link').forEach(function(el) {
    var href = el.getAttribute('href');
    if (href && href !== '/' && path.startsWith(href)) {
      el.classList.add('active');
    } else if (href === '/' && path === '/') {
      el.classList.add('active');
    }
  });
})();

// Dashboard bar chart fix: handle float values
(function() {
  document.querySelectorAll('.bar-bar').forEach(function(el) {
    if (el.style.height === '0px' || el.style.height === '') {
      el.style.height = '20px';
    }
  });
})();

// Coach: dynamic form auto-submit on select change (optional UX)
(function() {
  var dynamicForm = document.getElementById('dynamic-form');
  if (dynamicForm) {
    // no auto-submit; keep explicit button UX
  }
})();

// Check-in: prevent double submit
(function() {
  var checkinForm = document.querySelector('.checkin-form');
  if (checkinForm) {
    checkinForm.addEventListener('submit', function() {
      var btn = checkinForm.querySelector('button[type="submit"]');
      if (btn) {
        btn.disabled = true;
        btn.textContent = 'Saving…';
      }
    });
  }
})();
