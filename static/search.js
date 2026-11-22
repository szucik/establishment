// Person search box. Usage: <div data-person-search></div> anywhere on a page that loads this script.
// Results are built with textContent only, so person data can never inject markup.
(function () {
  const MAX_RESULTS = 8;
  const MIN_QUERY_LENGTH = 1;
  const MIN_STEMMED_WORD_LENGTH = 6;

  let personsPromise = null;

  function loadPersons() {
    personsPromise ??= fetch('/persons')
      .then((response) => (response.ok ? response.json() : []))
      .catch(() => []);
    return personsPromise;
  }

  function normalize(text) {
    return String(text ?? '').normalize('NFD').replace(/[̀-ͯ]/g, '')
      .replace(/ł/g, 'l').replace(/Ł/g, 'L').toLowerCase();
  }

  // Polish surnames change their ending with gender (Zieliński / Zielińska), so longer words match by stem.
  function stemOf(word) {
    return word.length >= MIN_STEMMED_WORD_LENGTH ? word.slice(0, -1) : word;
  }

  function matchesQuery(person, query) {
    const haystack = normalize(`${person.name} ${person.occupation} ${person.description} ${person.id}`);
    return query.split(/\s+/).every((word) => haystack.includes(stemOf(word)));
  }

  function profileUrlOf(person) {
    return `/static/person_view.html?id=${encodeURIComponent(person.id)}`;
  }

  function planetBackgroundOf(person) {
    if (!window.planetPaletteOf) return '#4a6a94';
    const [lit, body, shadow] = window.planetPaletteOf(person.id);
    return `radial-gradient(circle at 40% 38%, ${lit} 0%, ${body} 55%, ${shadow} 100%)`;
  }

  function initialsOf(name) {
    return (name || '?').split(/\s+/).filter(Boolean).slice(0, 2).map((word) => word[0].toUpperCase()).join('');
  }

  function resultItem(person, index) {
    const item = document.createElement('a');
    item.className = 'search-result';
    item.href = profileUrlOf(person);
    item.dataset.index = index;
    item.setAttribute('role', 'option');

    const planet = document.createElement('span');
    planet.className = 'search-planet';
    planet.style.background = planetBackgroundOf(person);
    planet.textContent = initialsOf(person.name);

    const text = document.createElement('span');
    text.className = 'search-text';
    const name = document.createElement('span');
    name.className = 'search-name';
    name.textContent = person.name;
    const occupation = document.createElement('span');
    occupation.className = 'search-occupation';
    occupation.textContent = person.occupation || '';
    text.append(name, occupation);

    item.append(planet, text);
    return item;
  }

  function emptyItem(query) {
    const item = document.createElement('div');
    item.className = 'search-empty';
    item.textContent = `Brak osób dla „${query}”`;
    return item;
  }

  function mountPersonSearch(host) {
    host.classList.add('person-search');
    const input = document.createElement('input');
    input.type = 'search';
    input.placeholder = 'Szukaj osoby…';
    input.autocomplete = 'off';
    input.setAttribute('aria-label', 'Szukaj osoby');
    const list = document.createElement('div');
    list.className = 'search-results';
    list.setAttribute('role', 'listbox');
    list.hidden = true;
    host.append(input, list);

    let results = [];
    let activeIndex = -1;

    const highlight = (index) => {
      activeIndex = index;
      list.querySelectorAll('.search-result').forEach((item) => {
        item.classList.toggle('active', Number(item.dataset.index) === index);
      });
    };

    const render = async () => {
      const query = normalize(input.value.trim());
      if (query.length < MIN_QUERY_LENGTH) {
        list.hidden = true;
        return;
      }
      const persons = await loadPersons();
      results = persons.filter((person) => matchesQuery(person, query)).slice(0, MAX_RESULTS);
      list.replaceChildren(...(results.length ? results.map(resultItem) : [emptyItem(input.value.trim())]));
      list.hidden = false;
      highlight(results.length ? 0 : -1);
    };

    input.addEventListener('focus', loadPersons);
    input.addEventListener('input', render);
    input.addEventListener('keydown', (event) => {
      if (event.key === 'ArrowDown' && results.length) {
        event.preventDefault();
        highlight((activeIndex + 1) % results.length);
      } else if (event.key === 'ArrowUp' && results.length) {
        event.preventDefault();
        highlight((activeIndex - 1 + results.length) % results.length);
      } else if (event.key === 'Enter' && results[activeIndex]) {
        window.location.href = profileUrlOf(results[activeIndex]);
      } else if (event.key === 'Escape') {
        list.hidden = true;
        input.blur();
      }
    });
    list.addEventListener('mousemove', (event) => {
      const item = event.target.closest('.search-result');
      if (item) highlight(Number(item.dataset.index));
    });
    document.addEventListener('click', (event) => {
      if (!host.contains(event.target)) list.hidden = true;
    });
  }

  function mountAll() {
    document.querySelectorAll('[data-person-search]').forEach(mountPersonSearch);
  }

  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', mountAll);
  else mountAll();
})();
