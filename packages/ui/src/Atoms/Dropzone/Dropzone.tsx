import { useCallback } from "react";
import { useDropzone } from "react-dropzone";

type Props = {};

export function Dropzone({}: Props) {
    const onDrop = useCallback((acceptedFiles) => {
        // Do something with the files
    }, []);
    const { getRootProps, getInputProps, isDragActive } = useDropzone({
        onDrop,
    });

    return (
        <div {...getRootProps()}>
            <input {...getInputProps()} />
            {isDragActive ? (
                <p>Drop the files here ...</p>
            ) : (
                <p>Drag 'n' drop some files here, or click to select files</p>
            )}
        </div>
    );
}
