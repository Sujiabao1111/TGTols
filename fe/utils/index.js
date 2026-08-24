export const formatDate = (date) => {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  const seconds = String(date.getSeconds()).padStart(2, '0');

  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
}

export const getLastSegment = (str) => {
  // 将字符串按/分割成数组（如"a/b/c"分割后是["a","b","c"]）
  const segments = str.split('/');
  // pop() 方法删除并返回数组的最后一个元素
  const lastSegment = segments.pop();
  // 处理特殊情况：如果最后一个字符是/（如"a/b/"），split后最后一个元素是空字符串
  return lastSegment || segments.pop() || '';
}

/**
 * 删除pathname中包含locale的部分（含对应的/）
 * @param {string} locale - 要删除的语言标识（如"cn"、"en"）
 * @param {string} pathname - 原始路径（如"/cn"、"/cn/home"、"/home/cn/about"）
 * @returns {string} 处理后的路径字符串
 */
export const removeLocaleFromPath = (locale, pathname) => {
  // 边界处理：如果locale为空或pathname为空，直接返回原pathname（去重/并trim）
  if (!locale || !pathname) {
    return pathname.replace(/\/+/g, '/').replace(/^\/|\/$/g, '');
  }

  // 正则表达式解释：
  // \/? 匹配0个或1个/（处理pathname可能以/开头的情况）
  // (^|\/) 匹配字符串开头或/（确保locale是独立的路径段）
  // ${locale} 匹配传入的locale值（如"cn"）
  // (\/|$) 匹配/或字符串结尾（确保locale是独立的路径段）
  // 全局匹配+忽略大小写（可选，根据需求调整）
  const regex = new RegExp(`(^|\\/)${escapeRegExp(locale)}(\\/|$)`, 'gi');

  // 第一步：替换匹配的locale相关部分为空字符串
  let result = pathname.replace(regex, '');
  // 第二步：去除多余的/（如"//home//about"变成"/home/about"）
  result = result.replace(/\/+/g, '/');
  // 第三步：去除开头和结尾的/（如"/home"变成"home"，""保持为空）
  result = result.replace(/^\/|\/$/g, '');

  return result;
}

/**
 * 辅助函数：转义正则表达式中的特殊字符（避免locale含特殊字符时出错）
 * @param {string} str - 要转义的字符串
 * @returns {string} 转义后的正则安全字符串
 */
function escapeRegExp(str) {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}
