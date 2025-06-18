## API Response

### Body - input: array
{
    model: string
    input: [
        {
            role: string
            type: string("message")
            content: [
                {
                    type: string("input_text")
                    text: string                
                },
                {
                    type: string("input_text")
                    text: string                
                },
            ]
            
        },        
    ]
}
