let tg, initData;

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

/**
 * Выполняет прямой переход на указанный URL с параметрами запроса.
 * @param {string} url - URL страницы, на которую нужно перейти
 * @param {string} urlQuery - Строка параметров запроса (без начального "?")
 */
function navigateWithParams(url, urlQuery) {
  let fullUrl = urlQuery ? `${url}?${urlQuery}` : url;
  window.location.href = fullUrl;
}
// main
document.addEventListener("DOMContentLoaded", function () {
  // Иницилизация телеграмм
  const result = initializeTg(5, 2000); // 5 попыток с интервалом в 2 секунды
  if (result) {
    console.log("Telegram Web App успешно инициализирован");
    tg = result.tg;
    initData = result.initData;
    // Дальнейшая логика с использованием tg и initData
  } else {
    console.error("Ошибка инициализации Telegram Web App");
    alert("Пожалуйста, запустите форму через телеграмм.");
    return;
  }

  console.log("initData");
  let urlQuery = "initData=" + encodeURIComponent(initData);

  navigateWithParams("/form", urlQuery)
    .then((result) => {
      if (!result.success) {
        alert("Navigation failed:", result.response);
        console.error("Navigation failed:", result.response);
      }
    })
    .catch((error) => {
      alert("Error:", error);
      console.error("Error:", error);
    });
});
