/**
 * Отправляет POST запрос по указанному url. В качестве тела запроса отправляет body
 * @param {string} url
 * @param {string} body
 * @returns
 */
function sendPostRequest(url, body) {
  var Request = false;

  if (window.XMLHttpRequest) {
    Request = new XMLHttpRequest();
  } else if (window.ActiveXObject) {
    try {
      Request = new ActiveXObject("Microsoft.XMLHTTP");
    } catch (CatchException) {
      try {
        Request = new ActiveXObject("Msxml2.XMLHTTP");
      } catch (CatchException2) {
        Request = false;
      }
    }
  }

  if (!Request) {
    return { success: false, response: "Невозможно создать XMLHttpRequest" };
  }

  try {
    Request.open("POST", url, false);
    Request.setRequestHeader("Content-Type", "application/json");

    // Преобразуем body в JSON, если это объект
    const jsonBody = typeof body === "object" ? JSON.stringify(body) : body;

    Request.send(jsonBody);

    if (Request.status === 200) {
      return { success: true, response: "" };
    } else {
      let response = Request.responseText;
      if (response === "") {
        response = `${Request.status} (${Request.statusText})`;
      }
      return { success: false, response: response };
    }
  } catch (error) {
    return { success: false, response: error.toString() };
  }
}

/**
 * Отправляет GET запрос по указанному url. В качестве параметров отправляет urlQuery.
 * Нербходимо самостоятельно заранее закодировать urlQuery. Указать надо без "?"
 * @param {string} url
 * @param {string} urlQuery
 * @returns
 */
function sendGetRequest(url, urlQuery) {
  var Request = false;

  if (window.XMLHttpRequest) {
    Request = new XMLHttpRequest();
  } else if (window.ActiveXObject) {
    try {
      Request = new ActiveXObject("Microsoft.XMLHTTP");
    } catch (CatchException) {
      try {
        Request = new ActiveXObject("Msxml2.XMLHTTP");
      } catch (CatchException2) {
        Request = false;
      }
    }
  }

  if (!Request) {
    return { success: false, response: "Невозможно создать XMLHttpRequest" };
  }

  try {
    // Изменяем эту часть, чтобы избежать двойного кодирования
    const fullUrl = urlQuery ? `${url}?${urlQuery}` : url;

    console.log("fullUrl    ", fullUrl);
    Request.open("GET", fullUrl, false);
    Request.send();

    if (Request.status === 200) {
      return { success: true, response: Request.responseText };
    } else {
      let response;
      switch (Request.status) {
        case 404:
          response = "404 (Not Found)";
          break;
        case 403:
          response = "403 (Forbidden)";
          break;
        case 500:
          response = "500 (Internal Server Error)";
          break;
        default:
          response = `${Request.status} (${
            Request.statusText || "Unknown Error"
          })`;
      }
      return { success: false, response: response };
    }
  } catch (error) {
    return { success: false, response: "Network Error" };
  }
}

/**
 * Преобразует все строковые значения в объекте (и его вложенных объектах) в соответствующие HTML элементы.
 * Если элемент с указанным id не найден, выводится предупреждение в консоль.
 *
 * @param {Object} objects - Объект, содержащий другие объекты, значения которых нужно преобразовать
 */
function convertIdsToElements(objects) {
  /**
   * Рекурсивно обрабатывает объект, заменяя строковые значения на HTML элементы
   *
   * @param {Object} obj - Объект для обработки
   */
  function processObject(obj) {
    for (let key in obj) {
      if (typeof obj[key] === "string") {
        const id = obj[key]; // Используем значение как id
        const element = document.getElementById(id);
        if (element) {
          obj[key] = element; // Заменяем строку на HTML элемент
        } else {
          console.warn(`Элемент с id "${id}" не найден для ключа "${key}"`);
        }
      } else if (typeof obj[key] === "object" && obj[key] !== null) {
        processObject(obj[key]); // Рекурсивный вызов для вложенных объектов
      }
    }
  }

  // Обрабатываем каждый объект в переданном objects
  for (let objName in objects) {
    processObject(objects[objName]);
  }
}

/**
 * setTg инициализирует Telegram Web App и возвращает объекты tg и initData
 * @param {number} [maxAttempts=3] Максимальное количество попыток инициализации
 * @param {number} [delay=1000] Задержка между попытками в миллисекундах
 * @returns {Object|null} Объект с свойствами tg и initData или null в случае неудачи
 */
function initializeTg(maxAttempts = 3, delay = 1000) {
  function sleep(ms) {
    const start = Date.now();
    while (Date.now() - start < ms) {}
  }

  for (let attempts = 1; attempts <= maxAttempts; attempts++) {
    if (window.Telegram && window.Telegram.WebApp) {
      const tg = window.Telegram.WebApp;
      const initData = tg.initData;

      // Проверяем, что initData не пустой и не равен ""
      if (!initData || initData === "") {
        console.warn(
          `Попытка ${attempts}: initData пуст или равен "". Возможно, приложение запущено не в Telegram.`
        );
      } else {
        // Вызываем tg.ready() для сообщения Telegram, что приложение готово
        tg.ready();

        // Добавим проверку на поддержку основных методов
        if (
          typeof tg.sendData !== "function" ||
          typeof tg.expand !== "function"
        ) {
          console.warn(
            `Попытка ${attempts}: Некоторые ожидаемые методы Telegram Web App отсутствуют.`
          );
        } else {
          console.log(
            `Telegram Web App успешно инициализирован (попытка ${attempts})`
          );
          return { tg, initData };
        }
      }
    } else {
      console.error(
        `Попытка ${attempts}: Telegram Web App не найден. Проверьте подключение библиотеки telegram-web-app.js.`
      );
    }

    if (attempts < maxAttempts) {
      console.log(`Повторная попытка через ${delay}мс...`);
      sleep(delay);
    }
  }

  console.error(
    `Не удалось инициализировать Telegram Web App после ${maxAttempts} попыток.`
  );
  return null;
}

// Глобальные переменные для отслеживания текущего предупреждения
let currentWarning = null; // Хранит текущий элемент предупреждения
let currentWarningTimeout = null; // Хранит таймер для автоматического скрытия предупреждения

/**
 * Показывает предупреждение рядом с целевым элементом.
 * @param {Element} targetElement - Элемент, рядом с которым нужно показать предупреждение.
 * @param {string} message - Текст предупреждения.
 * @param {string} position - Позиция предупреждения относительно целевого элемента (по умолчанию "top").
 * @param {string} percent - Горизонтальное смещение стрелочки предупреждения в процентах (по умолчанию "50%").
 * @param {number} duration - Продолжительность показа предупреждения в миллисекундах (по умолчанию 3000).
 */
function showWarning(
  targetElement,
  message,
  position = "top",
  percent = "50%",
  duration = 3000
) {
  console.log("showWarning called", {
    targetElement,
    message,
    position,
    percent,
    duration,
  });

  // Удаляем предыдущее предупреждение, если оно существует
  if (currentWarning) {
    currentWarning.remove();
  }

  // Очищаем предыдущий таймер, если он существует
  if (currentWarningTimeout) {
    clearTimeout(currentWarningTimeout);
  }

  // Проверяем, является ли targetElement допустимым DOM-элементом
  if (!targetElement || !(targetElement instanceof Element)) {
    console.error("Invalid targetElement");
    return;
  }

  // Создаем уникальный идентификатор для предупреждения
  const warningId = `warning-${
    targetElement.id || Math.random().toString(36).substr(2, 9)
  }`;

  // Создаем элемент предупреждения
  const warning = document.createElement("div");
  warning.id = warningId;
  warning.className = `warning warning-${position}`;
  warning.textContent = message;
  warning.style.visibility = "hidden";
  warning.style.opacity = "0";
  document.body.appendChild(warning);

  currentWarning = warning;

  console.log("Warning element created", warning);

  // Устанавливаем CSS-переменную для позиционирования псевдоэлемента
  document.documentElement.style.setProperty("--pseudo-left", percent);

  // Используем setTimeout, чтобы дать браузеру время на отрисовку предупреждения
  setTimeout(() => {
    // Получаем позицию целевого элемента
    const targetRect = targetElement.getBoundingClientRect();
    const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
    const scrollLeft =
      window.pageXOffset || document.documentElement.scrollLeft;

    // Позиционируем предупреждение
    warning.style.left = `${
      targetRect.left + scrollLeft + targetRect.width / 2
    }px`;
    warning.style.top = `${targetRect.top + scrollTop - 10}px`;
    warning.style.transform = "translate(-50%, -100%)";
    warning.style.visibility = "visible";

    console.log("Warning styles set", {
      position: warning.style.position,
      top: warning.style.top,
      left: warning.style.left,
      transform: warning.style.transform,
      zIndex: warning.style.zIndex,
      pseudoLeft: percent,
    });

    // Анимируем появление предупреждения
    requestAnimationFrame(() => {
      warning.style.opacity = "1";
      targetElement.scrollIntoView({ behavior: "smooth", block: "center" });
    });

    console.log("Warning position:", warning.getBoundingClientRect());
    console.log("Target position:", targetRect);

    // Устанавливаем таймер для автоматического скрытия предупреждения
    if (duration !== Infinity) {
      currentWarningTimeout = setTimeout(() => {
        hideWarning(targetElement);
      }, duration);
    }
  }, 0);
}

/**
 * Скрывает текущее предупреждение.
 * @param {Element} targetElement - Элемент, рядом с которым было показано предупреждение.
 */
function hideWarning(targetElement) {
  if (currentWarning) {
    console.log("Hiding warning", currentWarning);
    // Анимируем исчезновение предупреждения
    currentWarning.style.opacity = "0";
    setTimeout(() => {
      currentWarning.remove();
      currentWarning = null;
      console.log("Warning removed");
    }, 300);
  } else {
    console.log("No active warning to hide");
  }

  // Сбрасываем CSS-переменную для позиционирования псевдоэлемента
  document.documentElement.style.removeProperty("--pseudo-left");
}

document.addEventListener("touchend", function (event) {
  let button = document.getElementById("input-close-button");
  var target = event.target;
  var isInput = target.tagName === "INPUT" || target.tagName === "TEXTAREA";

  if (isInput) {
    button.style.display = "flex";
  }
});

//  Автоматическое изменение размера textarea
/* JavaScript */
document.addEventListener("DOMContentLoaded", function () {
  const textareas = document.querySelectorAll(".autoResizeTextarea");

  function autoResize() {
    // Сохраняем текущую позицию прокрутки
    const scrollTop = window.pageYOffset || document.documentElement.scrollTop;

    this.style.height = "auto";
    this.style.height = this.scrollHeight + "px";

    // Проверяем, не превышает ли высота максимальное значение
    const maxHeight = parseInt(window.getComputedStyle(this).maxHeight);
    if (this.scrollHeight > maxHeight) {
      this.style.height = maxHeight + "px";
      this.style.overflowY = "auto";
    } else {
      this.style.overflowY = "hidden";
    }

    // Восстанавливаем позицию прокрутки
    window.scrollTo(0, scrollTop);
  }

  function initTextarea(textarea) {
    // Устанавливаем начальную высоту равной одной строке
    const lineHeight = parseFloat(window.getComputedStyle(textarea).lineHeight);
    textarea.style.height = lineHeight + "px";

    // Добавляем обработчики событий
    textarea.addEventListener("input", autoResize);
    textarea.addEventListener("focus", autoResize);
  }

  // Инициализация после загрузки DOM
  textareas.forEach(initTextarea);

  // Обработка изменения размера окна
  window.addEventListener("resize", function () {
    textareas.forEach(autoResize);
  });
});
// Выпадающий список
document.addEventListener("DOMContentLoaded", function () {
  const customSelects = document.querySelectorAll(".custom-select");

  function initializeSelect(select) {
    const selected = select.querySelector(".select-selected");
    const search = select.querySelector(".select-search");
    const items = select.querySelector(".select-items");

    function handleItemClick(item) {
      const previousValue = selected.textContent;
      selected.textContent = item.textContent;
      closeDropdown(select);

      if (previousValue !== item.textContent) {
        if (selected.id === "teg" || selected.id === "replace-format") {
          const event = new CustomEvent(`${selected.id}Changed`, {
            detail: { value: item },
          });
          selected.dispatchEvent(event);
        }
      }
    }

    // Используем делегирование событий на уровне items
    items.addEventListener("click", function (e) {
      if (e.target.tagName === "LI") {
        handleItemClick(e.target);
      }
    });

    selected.addEventListener("click", function (e) {
      e.stopPropagation();
      closeAllDropdowns();
      items.style.display = "block";
      selected.style.display = "none";
      search.style.display = "block";
      search.focus();
      search.value = "";
      filterItems("");
    });

    search.addEventListener("input", function () {
      filterItems(this.value.toLowerCase());
    });

    function filterItems(filter) {
      const listItems = items.querySelectorAll("li");
      listItems.forEach((item) => {
        if (item.textContent.toLowerCase().indexOf(filter) > -1) {
          item.style.display = "";
        } else {
          item.style.display = "none";
        }
      });
    }
  }

  customSelects.forEach(initializeSelect);

  document.addEventListener("click", closeAllDropdowns);

  function closeDropdown(select) {
    const selected = select.querySelector(".select-selected");
    const search = select.querySelector(".select-search");
    const items = select.querySelector(".select-items");
    items.style.display = "none";
    selected.style.display = "block";
    search.style.display = "none";
    search.value = "";
    search.blur();
  }

  function closeAllDropdowns() {
    customSelects.forEach(closeDropdown);
  }

  window.initializeCustomSelect = initializeSelect;
});
