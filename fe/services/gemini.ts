import { GoogleGenAI, Chat, GenerateContentResponse } from "@google/genai";

// Initialize the Gemini AI client
// Note: In a real production environment, ensure process.env.API_KEY is set.
// If not set, the chat feature will gracefully handle the error or show a demo state.
const apiKey = process.env.API_KEY || '';
let ai: GoogleGenAI | null = null;

if (apiKey) {
  ai = new GoogleGenAI({ apiKey });
}

let chatSession: Chat | null = null;

export const initializeChat = async () => {
  if (!ai) return null;

  try {
    chatSession = ai.chats.create({
      model: 'gemini-2.5-flash',
      config: {
        systemInstruction: `You are Lucky Bear, the friendly and helpful mascot assistant for an online gaming platform called 'Lucky Bear Games'. 
        Your tone is enthusiastic, playful, and lucky. 
        You help users find games, explain rules for slots and blackjack, and provide general support.
        If asked about bonuses, mention the 'Welcome Bear Bundle' which gives 100% match up to $500.
        Keep responses concise (under 50 words) unless explaining a game rule detailedly.`,
      },
    });
    return chatSession;
  } catch (error) {
    console.error("Failed to initialize chat:", error);
    return null;
  }
};

export const sendMessageToGemini = async (message: string): Promise<string> => {
  if (!ai) {
    return "I'm currently offline (API Key missing). Please check back later!";
  }

  if (!chatSession) {
    await initializeChat();
  }

  if (!chatSession) {
     return "Lucky Bear is having trouble connecting right now. Try again?";
  }

  try {
    const response: GenerateContentResponse = await chatSession.sendMessage({ message });
    return response.text || "I didn't catch that, could you repeat it?";
  } catch (error) {
    console.error("Gemini API Error:", error);
    return "Oops! Something went wrong. Lucky Bear is confused.";
  }
};
