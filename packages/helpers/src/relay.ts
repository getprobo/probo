export function promisifyMutation<Response, Error, Args>(
    mutationFn: (
        args: {
            onCompleted?: (response: Response, error: Error) => void;
            onError?: (error: Error) => void;
        } & Args,
    ) => void,
): (
    args: {
        onCompleted?: (response: Response, error: Error) => void;
        onError?: (error: Error) => void;
    } & Args,
) => Promise<Response> {
    return (opts) => {
        return new Promise((resolve, reject) => {
            mutationFn({
                ...opts,
                onCompleted: (response, error) => {
                    if (opts.onCompleted) {
                        opts.onCompleted(response, error);
                    }
                    if (error) {
                        reject(error);
                    } else {
                        resolve(response);
                    }
                },
                onError: (error) => {
                    if (opts.onError) {
                        opts.onError(error);
                    }
                    reject(error);
                },
            });
        });
    };
}
