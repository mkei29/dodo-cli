import argparse
from openai import OpenAI
import sys

SYSTEM_PROMPT = """
You are a translator. You have been given a text to translate from one language to another.
Please translate the following text from "{input_language}" to "{output_language}".
Please prioritize natural expression over accuracy.

Input: A markdown document written in "{input_language}".
Output: A markdown document written in "{output_language}".

"""

def translate_text(text:str, input_language: str, output_language: str) -> str:
    if input_language == output_language:
        raise ValueError("Input and output languages should be different")

    client = OpenAI()
    chat_completion = client.chat.completions.create(
        messages=[
            {
                "role": "system",
                "content": SYSTEM_PROMPT.format(input_language=input_language, output_language=output_language),
            },
            { 
                "role": "user",
                "content": text
             }
        ],
        model="gpt-3.5-turbo",
    )

    if len(chat_completion.choices) != 1:
        raise RuntimeError("Unexpected number of messages in chat completion")
    return chat_completion.choices[0].message.content

    

def main():
    parser = argparse.ArgumentParser(description="Translate a text to another language")
    parser.add_argument("text", help="The text to translate")
    parser.add_argument("-i", "--input-language", help="The language to translate to", default="en")
    parser.add_argument("-o", "--output-language", help="The language to translate to", default="en")

    args = parser.parse_args()

    with open(args.text, "r") as f:
        text = f.read()
    translated = translate_text(text, args.input_language, args.output_language)
    sys.stdout.write(translated)



if __name__ == "__main__":
    main()
