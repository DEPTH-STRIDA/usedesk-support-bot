let tg, initData;

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
    showAlert("Пожалуйста, запустите форму через телеграмм", 15, true);
    return;
  }
  setButtonHanders();
});

function setButtonHanders() {
  let consoleArea = document.getElementById("console");
  let button0 = document.getElementById("update-teg-problems");

  button0.addEventListener("click", async function () {
    button0.disabled = true;
    button0.innerHTML = "";
    document.documentElement.style.setProperty("--main-spin-width", "5vw");
    document.documentElement.style.setProperty("--main-spin-heigt", "5vw");
    button0.classList.add("spin-main");

    const url = "/admin-command";
    let urlQuery =
      "initData=" +
      encodeURIComponent(initData) +
      "&command=update-select-data";

    try {
      const result = await sendGetRequestAsync(url, urlQuery);

      if (result.success) {
        console.log("Тело ответа:", result.response);
        let oldValue = consoleArea.innerHTML;
        let newValue = result.response + "\n" + oldValue;
        consoleArea.innerHTML = newValue;
      } else {
        console.error("Запрос выполнен, но вернул ошибку:", result.response);
        let oldValue = consoleArea.innerHTML;
        let newValue = result.response + "\n" + oldValue;
        consoleArea.innerHTML = newValue;
      }
    } catch (error) {
      console.error("Произошла ошибка при выполнении запроса:", error);
      showAlert("Ошибка: " + error, 5, false);
    } finally {
      button0.disabled = false;
      button0.innerHTML = "Отправить";
      button0.classList.remove("spin-main");
    }
  });
}

/**
 * Отправляет асинхронный GET запрос по указанному url. В качестве параметров отправляет urlQuery.
 * Необходимо самостоятельно заранее закодировать urlQuery. Указать надо без "?"
 * @param {string} url
 * @param {string} urlQuery
 * @returns {Promise<{success: boolean, response: string}>}
 */
function sendGetRequestAsync(url, urlQuery) {
  return new Promise((resolve, reject) => {
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
      reject(new Error("Невозможно создать XMLHttpRequest"));
      return;
    }

    const fullUrl = urlQuery ? `${url}?${urlQuery}` : url;
    console.log("fullUrl    ", fullUrl);

    Request.open("GET", fullUrl, true); // Устанавливаем третий параметр в true для асинхронного запроса

    Request.onreadystatechange = function () {
      if (Request.readyState === 4) {
        // Запрос завершен
        if (Request.status === 200) {
          resolve({ success: true, response: Request.responseText });
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
          resolve({ success: false, response: response });
        }
      }
    };

    Request.onerror = function () {
      reject(new Error("Network Error"));
    };

    try {
      Request.send();
    } catch (error) {
      reject(error);
    }
  });
}
