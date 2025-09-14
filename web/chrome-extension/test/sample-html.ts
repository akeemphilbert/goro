/**
 * Sample HTML with microformats for testing
 */

export const SAMPLE_HCARD_HTML = `
<div class="h-card">
  <img class="u-photo" src="https://example.com/photo.jpg" alt="John Doe" />
  <h1 class="p-name">John Doe</h1>
  <p class="p-org">Acme Corporation</p>
  <p class="p-job-title">Software Engineer</p>
  <a class="u-email" href="mailto:john@example.com">john@example.com</a>
  <a class="u-url" href="https://johndoe.com">Personal Website</a>
  <p class="p-tel">+1-555-123-4567</p>
  <p class="p-note">Passionate software engineer with 10 years of experience.</p>
  <div class="p-adr h-adr">
    <span class="p-street-address">123 Main St</span>
    <span class="p-locality">Anytown</span>
    <span class="p-region">CA</span>
    <span class="p-postal-code">12345</span>
  </div>
</div>
`;

export const SAMPLE_HEVENT_HTML = `
<div class="h-event">
  <h1 class="p-name">Team Meeting</h1>
  <p class="p-summary">Weekly team synchronization meeting</p>
  <time class="dt-start" datetime="2023-12-01T10:00:00">December 1, 2023 at 10:00 AM</time>
  <time class="dt-end" datetime="2023-12-01T11:00:00">11:00 AM</time>
  <p class="p-location">Conference Room A</p>
  <p class="p-description">
    Weekly meeting to discuss project progress, blockers, and upcoming tasks.
    All team members are expected to attend.
  </p>
  <a class="u-url" href="https://example.com/meetings/team-weekly">Meeting Details</a>
  <p class="p-category">work</p>
  <p class="p-category">meeting</p>
</div>
`;

export const SAMPLE_HPRODUCT_HTML = `
<div class="h-product">
  <h1 class="p-name">Awesome Widget Pro</h1>
  <img class="u-photo" src="https://example.com/widget-pro.jpg" alt="Awesome Widget Pro" />
  <p class="p-brand">Acme Corporation</p>
  <p class="p-category">Electronics</p>
  <p class="p-category">Gadgets</p>
  <p class="p-description">
    The most advanced widget on the market. Features include wireless connectivity,
    AI-powered automation, and a sleek modern design.
  </p>
  <p class="p-price">$299.99</p>
  <p class="p-identifier">WIDGET-PRO-2023</p>
  <a class="u-url" href="https://example.com/products/widget-pro">Product Page</a>
  <div class="p-review h-review">
    <p class="p-rating">5</p>
    <p class="p-description">Amazing product, highly recommended!</p>
  </div>
</div>
`;

export const SAMPLE_HRECIPE_HTML = `
<div class="h-recipe">
  <h1 class="p-name">Chocolate Chip Cookies</h1>
  <img class="u-photo" src="https://example.com/cookies.jpg" alt="Chocolate Chip Cookies" />
  <p class="p-summary">Classic homemade chocolate chip cookies</p>
  <p class="p-author h-card">
    <span class="p-name">Chef Jane</span>
  </p>
  <time class="dt-published" datetime="2023-12-01">December 1, 2023</time>
  
  <h2>Ingredients:</h2>
  <ul>
    <li class="p-ingredient">2 cups all-purpose flour</li>
    <li class="p-ingredient">1 cup granulated sugar</li>
    <li class="p-ingredient">1/2 cup brown sugar</li>
    <li class="p-ingredient">1/2 cup butter, softened</li>
    <li class="p-ingredient">2 large eggs</li>
    <li class="p-ingredient">1 tsp vanilla extract</li>
    <li class="p-ingredient">1 cup chocolate chips</li>
  </ul>
  
  <h2>Instructions:</h2>
  <ol>
    <li class="p-instructions">Preheat oven to 350°F (175°C)</li>
    <li class="p-instructions">Mix dry ingredients in a large bowl</li>
    <li class="p-instructions">Cream butter and sugars, add eggs and vanilla</li>
    <li class="p-instructions">Combine wet and dry ingredients, fold in chocolate chips</li>
    <li class="p-instructions">Drop spoonfuls on baking sheet</li>
    <li class="p-instructions">Bake for 10-12 minutes until golden brown</li>
  </ol>
  
  <p class="p-yield">Makes 24 cookies</p>
  <p class="p-duration">PT45M</p>
  <p class="p-category">dessert</p>
  <p class="p-category">baking</p>
</div>
`;

export const SAMPLE_HENTRY_HTML = `
<article class="h-entry">
  <h1 class="p-name">Getting Started with Microformats</h1>
  <p class="p-summary">
    An introduction to microformats and how they can improve your website's semantic markup.
  </p>
  
  <div class="p-author h-card">
    <img class="u-photo" src="https://example.com/author.jpg" alt="Jane Blogger" />
    <span class="p-name">Jane Blogger</span>
    <a class="u-url" href="https://janeblogger.com">janeblogger.com</a>
  </div>
  
  <time class="dt-published" datetime="2023-12-01T09:00:00">December 1, 2023</time>
  <time class="dt-updated" datetime="2023-12-01T10:30:00">Updated: 10:30 AM</time>
  
  <div class="e-content">
    <p>
      Microformats are a simple way to add semantic meaning to your HTML markup.
      They allow machines to understand the structure and meaning of your content,
      making it easier for search engines, social media platforms, and other tools
      to process and display your information correctly.
    </p>
    
    <p>
      In this post, we'll explore the basics of microformats and show you how to
      implement them on your website. We'll cover h-card for contact information,
      h-event for events, and h-entry for blog posts.
    </p>
  </div>
  
  <a class="u-url" href="https://blog.example.com/microformats-intro">Permalink</a>
  <p class="p-category">web development</p>
  <p class="p-category">semantic markup</p>
  <p class="p-category">microformats</p>
</article>
`;

export const SAMPLE_MIXED_HTML = `
<html>
<head>
  <title>Sample Page with Multiple Microformats</title>
</head>
<body>
  <header>
    ${SAMPLE_HCARD_HTML}
  </header>
  
  <main>
    ${SAMPLE_HENTRY_HTML}
    
    <section>
      ${SAMPLE_HEVENT_HTML}
    </section>
    
    <aside>
      ${SAMPLE_HPRODUCT_HTML}
      ${SAMPLE_HRECIPE_HTML}
    </aside>
  </main>
</body>
</html>
`;

export const SAMPLE_NO_MICROFORMATS_HTML = `
<html>
<head>
  <title>Page Without Microformats</title>
</head>
<body>
  <h1>Regular HTML Page</h1>
  <p>This page contains no microformats.</p>
  <div class="regular-content">
    <p>Just regular HTML content here.</p>
  </div>
</body>
</html>
`;

export const SAMPLE_INVALID_MICROFORMATS_HTML = `
<div class="h-card">
  <!-- Missing required properties -->
</div>
<div class="h-event">
  <span class="p-name"></span> <!-- Empty name -->
</div>
<div class="not-a-microformat">
  <span class="p-name">This won't be parsed</span>
</div>
`;